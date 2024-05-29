package chunkserver

import (
	"fmt"
	"glfs/common"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync/atomic"
)

// TODO: persist ChunkServer state
type ChunkServer struct {
	Id          uint32
	Address     string
	Chunks      map[uint64]*Chunk
	Updates     map[uint64]*Update
	UpdateIndex atomic.Uint64
}

type Chunk struct {
	ChunkHandle uint64
	Version     uint64
}

type Update struct {
	ChunkHandle uint64
	UpdateId    uint64
	Offset      uint64
	Data        []byte
}

func (t *ChunkServer) Ping(args *common.PingArgs, reply *bool) error {
	*reply = true
	return nil
}

func (t *ChunkServer) Create(args *common.CreateFileArgsChunk, reply *bool) error {
	log.Printf("Received Chunk.Create call with chunkHandle %v and chunkSize %v", args.ChunkHandle, len(args.Content))

	filePath := t.getChunkFilePath(args.ChunkHandle)
	err := os.WriteFile(filePath, args.Content, 0644)
	common.Check(err)

	t.Chunks[args.ChunkHandle] = &Chunk{
		ChunkHandle: args.ChunkHandle,
		Version:     0,
	}

	log.Printf("Successfully saved local %v", filePath)

	*reply = true
	return nil
}

func (t *ChunkServer) Delete(args *common.DeleteFileArgsChunk, reply *bool) error {
	log.Printf("Received Chunk.Delete call with chunkHandle %v", args.ChunkHandle)
	filePath := t.getChunkFilePath(args.ChunkHandle)
	err := os.Remove(filePath)
	common.Check(err)

	delete(t.Chunks, args.ChunkHandle)

	log.Printf("Successfully deleted local %v", filePath)

	*reply = true
	return nil
}

func (t *ChunkServer) Read(args *common.ReadFileArgsChunk, reply *common.ReadFileReplyChunk) error {
	log.Printf("Received Chunk.Read call with chunkHandle %v", args.ChunkHandle)

	filePath := t.getChunkFilePath(args.ChunkHandle)
	content, err := os.ReadFile(filePath)
	common.Check(err)

	log.Printf("Successfully read from local %v", filePath)

	reply.Content = content
	return nil
}

func (t *ChunkServer) Write(args *common.WriteArgsChunk, reply *common.WriteReplyChunk) error {
	// save the updates for now in memory, and wait for the CommitWrite message
	// atomically increase update index
	updateIndex := t.UpdateIndex.Add(1)
	t.Updates[updateIndex] = &Update{
		ChunkHandle: args.ChunkHandle,
		UpdateId:    updateIndex,
		Offset:      args.Offset,
		Data:        args.Data,
	}

	reply.UpdateId = updateIndex

	return nil
}

func (t *ChunkServer) CommitWrite(args *common.CommitWriteArgsChunk, reply *bool) error {
	t.commitWrite(args.UpdateId)

	// if primary, send commit messages to all other replicas after commiting the update on own state
	if args.IsPrimary {
		t.commitWritesReplicas(args.Replicas)
	}

	*reply = true

	return nil
}

func InitializeChunkServer(idStr *string) {
	// Init chunk server
	chunk := new(ChunkServer)
	id, _ := strconv.Atoi(*idStr)
	chunk.Id = uint32(id)
	chunk.Address = common.GetChunkServerAddress(chunk.Id)
	chunk.Chunks = map[uint64]*Chunk{}
	chunk.Updates = map[uint64]*Update{}

	// Create a directory for holding all chunks with chunkHandle as file names
	dirPath := common.GetTmpPath(fmt.Sprintf("chunk/%v", chunk.Id), "")
	err := os.MkdirAll(dirPath, os.ModePerm)
	common.Check(err)

	// ping master and join cluster
	masterClient, err := rpc.DialHTTP("tcp", common.GetMasterServerAddress())
	if err != nil {
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	args := &common.PingArgs{
		Id:      chunk.Id,
		Address: chunk.Address,
	}
	var reply bool
	err = masterClient.Call("MasterServer.Ping", args, &reply)
	if err != nil {
		log.Fatal(err)
	}

	// start serving
	rpc.Register(chunk)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", chunk.Address)
	if err != nil {
		log.Fatal(err)
	}
	http.Serve(l, nil)
}

func (t *ChunkServer) getChunkFilePath(chunkHandle uint64) string {
	return common.GetTmpPath(fmt.Sprintf("chunk/%v", t.Id), fmt.Sprint(chunkHandle))
}

func (t *ChunkServer) commitWritesReplicas(replicas map[string]uint64) {
	for addr, updateId := range replicas {
		chunkClient, err := rpc.DialHTTP("tcp", addr)
		if err != nil {
			log.Fatal("dialing:", err)
		}
		// Synchronous call
		args := &common.CommitWriteArgsChunk{
			IsPrimary: false,
			UpdateId:  updateId,
		}
		var reply bool
		err = chunkClient.Call("ChunkServer.CommitWrite", args, &reply)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (t *ChunkServer) commitWrite(updateId uint64) {
	update := t.Updates[updateId]

	// Increament version of this chunk
	// t.Chunks[update.ChunkHandle].Version++

	// Open the file for writing
	filePath := t.getChunkFilePath(update.ChunkHandle)
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	common.Check(err)

	defer file.Close()

	// Write at offset
	_, err = file.WriteAt(update.Data, int64(update.Offset))
	common.Check(err)
}
