package client

import (
	"glfs/common"
	"log"
	"os"
	"sync"
)

// Client will live along with application, serving as a library
// No need for RPC calls.
type GLFSClient struct {
	masterAddress string
}

// This method will be exported and called directly by application.
func (t *GLFSClient) Create(filepath string) bool {
	// handle error cases
	if len(t.masterAddress) == 0 {
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

	err = common.DialAndCall("MasterServer.Create", t.masterAddress, masterArgs, &reply)
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

		var wg sync.WaitGroup
		wg.Add(1)
		go t.sendFileToChunkServer(chunkInfo.PrimaryLocation, args, &wg)
		for _, replicaLocation := range chunkInfo.ReplicaLocations {
			wg.Add(1)
			go t.sendFileToChunkServer(replicaLocation, args, &wg)
		}

		wg.Wait()
	}
	return true
}

func (t *GLFSClient) Delete(filename string) bool {
	masterArgs := &common.DeleteFileArgsMaster{
		FileName: filename,
	}
	log.Printf("Calling Master.Delete with args %v", *masterArgs)

	var reply bool

	err := common.DialAndCall("MasterServer.Delete", t.masterAddress, masterArgs, &reply)
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

	err := common.DialAndCall("MasterServer.Read", t.masterAddress, masterArgs, &reply)
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
	var wg sync.WaitGroup
	for i, chunkInfo := range reply.Chunks {
		args := &common.ReadFileArgsChunk{
			ChunkHandle: chunkInfo.ChunkHandle,
		}
		content := make([]byte, common.ChunkSize)
		wg.Add(2)
		go t.readFileFromChunkServer(chunkInfo.PrimaryLocation, args, &content, &wg)
		offset := i * common.ChunkSize
		go func() {
			file.WriteAt(content, int64(offset))
			wg.Done()
		}()
	}

	wg.Wait()

	return nil
}

func (t *GLFSClient) Write(filename string, offset uint64, data []byte) {
	// First get chunkHandles and chunkLocations from master
	masterArgs := &common.GetPrimaryArgsMaster{
		FileName:   filename,
		ChunkIndex: offset / common.ChunkSize,
	}
	log.Printf("Calling Master.GetPrimary with args %v", *masterArgs)

	var reply common.ClientChunkInfo

	err := common.DialAndCall("MasterServer.GetPrimary", t.masterAddress, masterArgs, &reply)
	if err != nil {
		log.Fatal(err)
	}

	primaryAddress := reply.PrimaryLocation

	log.Printf("Got reply from master for GetPrimary: %v", reply)

	// push update to all chunkservers
	replicas := make(map[string]uint64, len(reply.ReplicaLocations))
	var primaryUpdateId uint64
	for _, addr := range reply.ReplicaLocations {
		args := &common.WriteArgsChunk{
			ChunkHandle: reply.ChunkHandle,
			Offset:      offset,
			Data:        data,
		}
		// skip adding master to replicas
		if addr != primaryAddress {
			replicas[addr] = t.sendUpdateToChunkServer(addr, args)
		} else {
			primaryUpdateId = t.sendUpdateToChunkServer(addr, args)
		}
	}

	// send commitWrite message to primary chunk server only
	args := &common.CommitWriteArgsChunk{
		IsPrimary: true,
		UpdateId:  primaryUpdateId,
		Replicas:  replicas,
	}
	commitReply := t.commitWriteAtPrimary(primaryAddress, args)
	log.Printf("Got commitReply from primary", commitReply)
}

// Returns updateId used to identify this update the chunkServer
func (t *GLFSClient) sendUpdateToChunkServer(location string, args *common.WriteArgsChunk) uint64 {
	// This is the changeId used to reference this change at chunkserver
	// later in commitWrite call, this will be used to apply this update.
	reply := &common.WriteReplyChunk{}

	log.Printf("Writing chunk chunkServer at %v with args %v", location, args)
	err := common.DialAndCall("ChunkServer.Write", location, args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received WriteReplyChunk from chunkServer at %v and reply %v", location, reply)

	return reply.UpdateId
}

// Returns updateId used to identify this update the chunkServer
func (t *GLFSClient) commitWriteAtPrimary(primaryAddr string, args *common.CommitWriteArgsChunk) bool {
	// This is the changeId used to reference this change at chunkserver
	// later in commitWrite call, this will be used to apply this update.
	var reply bool

	log.Printf("CommitWrite at %v with args %v", primaryAddr, args)
	err := common.DialAndCall("ChunkServer.CommitWrite", primaryAddr, args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received CommitWrite from primary chunkServer at %v and reply %v", primaryAddr, reply)

	return reply
}

func (t *GLFSClient) sendFileToChunkServer(location string, args *common.CreateFileArgsChunk, wg *sync.WaitGroup) {
	var reply bool

	log.Printf("Uploading file to chunkServer at %v", location)
	err := common.DialAndCall("ChunkServer.Create", location, args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received FileUploadReply from chunkServer at %v and reply %v", location, reply)

	wg.Done()
}

func (t *GLFSClient) readFileFromChunkServer(location string, args *common.ReadFileArgsChunk, content *[]byte, wg *sync.WaitGroup) {
	var reply common.ReadFileReplyChunk

	log.Printf("Downloading file from chunkServer at %v", location)
	err := common.DialAndCall("ChunkServer.Read", location, args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received FileDownloadReply from chunkServer at %v and content length %v", location, len(reply.Content))

	*content = reply.Content

	wg.Done()
}

// Initialize client with default configs
func (t *GLFSClient) Initialize() {
	t.masterAddress = common.GetMasterServerAddress()
}
