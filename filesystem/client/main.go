package main

import (
	"fmt"
	"glfs/common"
	"log"
	"net/rpc"
	"os"
)

// Client will live along with application, serving as a library
// No need for RPC calls.
type GLFSClient struct {
	masterClient *rpc.Client
}

// This method will be exported and called directly by application.
func (t *GLFSClient) Create(filepath string) bool {
	// handle error cases
	if t.masterClient == nil {
		log.Fatal("Connection to master server not initialized, please call Initialize() or manually create connection to master.")
	}

	fileInfo, err := os.Stat(filepath)
	if err != nil {
		log.Fatal("Cannot get file info, error: ", err)
	}

	// Call master's create to get back mapping of ChunkHandle -> ChunkServerAddr
	fileName, fileSize := fileInfo.Name(), fileInfo.Size()
	log.Printf("Found file with fileName: %v, fileSize: %v", fileName, fileSize)

	numberOfChunks := fileSize / common.ChunkSize

	args := &common.CreateFileArgs{
		FileName:       fileName,
		NumberOfChunks: uint8(numberOfChunks),
	}
	log.Printf("Calling Master.Create with args %v", *args)

	var reply common.CreateFileReply
	reply.ChunkMap = make(map[uint64]string)

	err = t.masterClient.Call("MasterServer.Create", args, &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Got reply from master for createFile: ", reply)

	for chunkHandle, chunkServerAddr := range reply.ChunkMap {
		// TODO: Call chunkservers with handle and chunks
	}
	return true
}

// Initialize client with default configs
func (t *GLFSClient) Initialize() {
	// connect to master server using tcp
	masterClient, err := rpc.DialHTTP("tcp", common.GetMasterServerAddress())
	if err != nil {
		log.Fatal("Connecting to master client failed: ", err)
	}
	// save the connection
	t.masterClient = masterClient
}

func main() {
	// test code for now
	client := GLFSClient{}
	client.Initialize()
	client.Create("/home/mutony/Projects/glfs/filesystem/tmp/test_0.dat")
}
