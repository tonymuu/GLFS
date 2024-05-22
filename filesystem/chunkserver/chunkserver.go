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

type ChunkServer struct {
	Id      uint8
	Address string
}

func (t *ChunkServer) Ping(args *common.PingArgs, reply *bool) error {
	*reply = true
	return nil
}

func (t *ChunkServer) Create(args *common.CreateFileArgsChunk, reply *bool) error {
	log.Printf("Received Chunk.Create call with chunkHandle %v and chunkSize %v", args.ChunkHandle, len(args.Content))

	filePath := common.GetTmpPath("chunk", fmt.Sprint(args.ChunkHandle))
	err := os.WriteFile(filePath, args.Content, 0644)
	common.Check(err)

	log.Printf("Successfully saved local %v", filePath)

	*reply = true
	return nil
}

func InitializeChunkServer(idStr *string) {
	// Init chunk server
	chunk := new(ChunkServer)
	id, _ := strconv.Atoi(*idStr)
	chunk.Id = uint8(id)
	chunk.Address = common.GetChunkServerAddress(chunk.Id)

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
