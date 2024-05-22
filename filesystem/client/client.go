package client

import (
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

	fileName, fileSize := fileInfo.Name(), fileInfo.Size()
	log.Printf("Found file with fileName: %v, fileSize: %v", fileName, fileSize)

	// Call master's create to get back mapping of ChunkHandle -> ChunkServerAddr
	numberOfChunks := fileSize / common.ChunkSize

	masterArgs := &common.CreateFileArgsMaster{
		FileName:       fileName,
		NumberOfChunks: uint8(numberOfChunks),
	}
	log.Printf("Calling Master.Create with args %v", *masterArgs)

	var reply common.CreateFileReplyMaster
	reply.ChunkMap = make(map[uint8]*common.ClientChunkInfo)

	err = t.masterClient.Call("MasterServer.Create", masterArgs, &reply)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got reply from master for createFile: ", reply)

	chunk := make([]byte, common.ChunkSize)
	file, _ := os.Open(filepath)
	defer file.Close()
	for chunkIndex, chunkInfo := range reply.ChunkMap {
		// Chunk the files into chunks of fixed size
		// read the current chunk, based on chunkIndex and chunkSize
		file.ReadAt(chunk, int64(chunkIndex)*common.ChunkSize)

		// Call chunkservers with handle and chunks
		args := &common.CreateFileArgsChunk{
			ChunkHandle: chunkInfo.ChunkHandle,
			Content:     chunk,
		}

		t.sendFileToChunkServer(chunkInfo.Location, args)
	}
	return true
}

func (t *GLFSClient) sendFileToChunkServer(location string, args *common.CreateFileArgsChunk) {
	// connect to master server using tcp
	chunkClient, err := rpc.DialHTTP("tcp", location)
	if err != nil {
		log.Fatal("Connecting to chunk client failed: ", err)
	}

	var reply bool

	log.Printf("Uploading file to chunkServer at %v", location)
	err = chunkClient.Call("ChunkServer.Create", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received FileUploadReply from chunkServer at %v and reply %v", location, reply)
}

// Initialize client with default configs
func (t *GLFSClient) Initialize() {
	// connect to master server using tcp
	masterClient, err := rpc.DialHTTP("tcp", common.GetMasterServerAddress())
	if err != nil {
		log.Fatal("Connecting to master client failed: ", err)
	}
	// persist the connection
	t.masterClient = masterClient
}

func InitializeClient() {
	// test code for now
	client := GLFSClient{}
	client.Initialize()
	client.Create(common.GetTmpPath("client", "test_0.dat"))
}
