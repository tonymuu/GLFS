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
		NumberOfChunks: uint32(numberOfChunks),
	}
	log.Printf("Calling Master.Create with args %v", *masterArgs)

	var reply common.CreateFileReplyMaster
	reply.ChunkMap = make(map[uint32]*common.ClientChunkInfo)

	err = t.masterClient.Call("MasterServer.Create", masterArgs, &reply)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got reply from master for createFile: %v", reply)

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

		t.sendFileToChunkServer(chunkInfo.PrimaryLocation, args)

		for _, replicaLocation := range chunkInfo.ReplicaLocations {
			t.sendFileToChunkServer(replicaLocation, args)
		}
	}
	return true
}

func (t *GLFSClient) Delete(filename string) bool {
	masterArgs := &common.DeleteFileArgsMaster{
		FileName: filename,
	}
	log.Printf("Calling Master.Delete with args %v", *masterArgs)

	var reply bool

	err := t.masterClient.Call("MasterServer.Delete", masterArgs, &reply)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got reply from master for delete: %v", reply)
	return reply
}

func (t *GLFSClient) Read(filename string, outputPath string) []byte {
	// First get chunkHandles and chunkLocations from master
	masterArgs := &common.ReadFileArgsMaster{
		FileName: filename,
	}
	log.Printf("Calling Master.Read with args %v", *masterArgs)

	var reply common.ReadFileReplyMaster

	err := t.masterClient.Call("MasterServer.Read", masterArgs, &reply)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got reply from master for read: %v", reply)

	// open the file and start appending in order
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Error creating output file at %v", outputPath)
	}
	defer file.Close()

	// Then for each chunk handle/location, get chunk bytes from chunkServers
	for i, chunkInfo := range reply.Chunks {
		args := &common.ReadFileArgsChunk{
			ChunkHandle: chunkInfo.ChunkHandle,
		}
		content := t.readFileFromChunkServer(chunkInfo.PrimaryLocation, args)
		offset := i * common.ChunkSize
		file.WriteAt(content, int64(offset))
	}

	return nil
}

func (t *GLFSClient) Write(filename string, offset uint64, data []byte) {
	// First get chunkHandles and chunkLocations from master
	masterArgs := &common.GetPrimaryArgsMaster{
		FileName:   filename,
		ChunkIndex: offset / common.ChunkSize,
	}
	log.Printf("Calling Master.GetPrimary with args %v", *masterArgs)

	var reply common.GetPrimaryReplyMaster

	err := t.masterClient.Call("MasterServer.GetPrimary", masterArgs, &reply)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got reply from master for GetPrimary: %v", reply)

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

func (t *GLFSClient) readFileFromChunkServer(location string, args *common.ReadFileArgsChunk) []byte {
	// connect to master server using tcp
	chunkClient, err := rpc.DialHTTP("tcp", location)
	if err != nil {
		log.Fatal("Connecting to chunk client failed: ", err)
	}

	var reply common.ReadFileReplyChunk

	log.Printf("Downloading file from chunkServer at %v", location)
	err = chunkClient.Call("ChunkServer.Read", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received FileDownloadReply from chunkServer at %v and content length", location, len(reply.Content))

	return reply.Content
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
}
