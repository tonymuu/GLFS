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
)

// TODO: persist ChunkServer state
type ChunkServer struct {
	Id      uint32
	Address string
	Chunks  map[uint64]*Chunk
}

type Chunk struct {
	ChunkHandle uint64
	Version     uint64
}

func (t *ChunkServer) Ping(args *common.PingArgs, reply *bool) error {
	*reply = true
	return nil
}

func (t *ChunkServer) Create(args *common.CreateFileArgsChunk, reply *bool) error {
	log.Printf("Received Chunk.Create call with chunkHandle %v and chunkSize %v", args.ChunkHandle, len(args.Content))

	filePath := common.GetTmpPath(fmt.Sprintf("chunk/%v", t.Id), fmt.Sprint(args.ChunkHandle))
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

func (t *ChunkServer) Read(args *common.ReadFileArgsChunk, reply *common.ReadFileReplyChunk) error {
	log.Printf("Received Chunk.Read call with chunkHandle %v", args.ChunkHandle)

	filePath := common.GetTmpPath(fmt.Sprintf("chunk/%v", t.Id), fmt.Sprint(args.ChunkHandle))
	content, err := os.ReadFile(filePath)
	common.Check(err)

	log.Printf("Successfully read from local %v", filePath)

	reply.Content = content
	return nil
}

func InitializeChunkServer(idStr *string) {
	// Init chunk server
	chunk := new(ChunkServer)
	id, _ := strconv.Atoi(*idStr)
	chunk.Id = uint32(id)
	chunk.Address = common.GetChunkServerAddress(chunk.Id)
	chunk.Chunks = map[uint64]*Chunk{}

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
