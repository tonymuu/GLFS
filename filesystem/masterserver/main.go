package masterserver

import (
	"fmt"
	"glfs/common"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type MasterServer struct {
	// manages chunk server information
	ChunkServers map[uint8]*ChunkServerMetadata

	// maps from filename to a list of chunk handles
	FileMetadata map[string]*FileMetadata

	// maps from chunkhandle to chunk metadata (location, expiration, etc.)
	ChunkMetadata map[common.ChunkHandle]*ChunkMetadata
}

type ChunkServerMetadata struct {
	ServerId          uint8
	ServerAddress     string
	TimeStampLastPing int64
}

type FileMetadata struct {
	FileName     string
	ChunkHandles *[]common.ChunkHandle
}

type ChunkMetadata struct {
	Location string
}

func (t *MasterServer) Ping(args *common.PingArgs, reply *bool) error {
	log.Printf("Master.Ping called with args %v", args)

	t.ChunkServers[args.Id] = &ChunkServerMetadata{
		ServerId:          args.Id,
		ServerAddress:     args.Address,
		TimeStampLastPing: time.Now().Unix(),
	}

	log.Printf("Updated ChunkServers info. New state: %v", t.ChunkServers)

	// Expired chunk servers are presumed dead, so we remove them.
	t.removeExpiredChunkServers()

	*reply = true
	return nil
}

// RPC method on the MasterServer used to create a file.
func (t *MasterServer) Create(args *common.CreateFileArgsMaster, reply *common.CreateFileReplyMaster) error {
	log.Printf("Received Master.Create call with args %v", args)

	// chunkId := uint8(0)
	reply.ChunkMap = map[uint8]*common.ClientChunkInfo{}

	chunkHandles := make([]common.ChunkHandle, args.NumberOfChunks)
	t.FileMetadata[args.FileName] = &FileMetadata{
		FileName:     args.FileName,
		ChunkHandles: &chunkHandles,
	}

	chunkServerIds := make([]uint8, len(t.ChunkServers))
	i := 0
	for chunkServerId := range t.ChunkServers {
		chunkServerIds[i] = chunkServerId
		i++
	}

	for chunkId := range args.NumberOfChunks {
		// Get chunkServer address
		chunkServer := t.ChunkServers[mapChunkIdToChunkServerIndex(chunkId, chunkServerIds)]
		chunkLocation := chunkServer.ServerAddress

		// compute chunkHandle
		chunkName := fmt.Sprintf("%v-%v", args.FileName, chunkId)
		chunkHandle := getChunkHandle(chunkName)

		// Save this information
		t.ChunkMetadata[chunkHandle] = &ChunkMetadata{}
		t.ChunkMetadata[chunkHandle].Location = chunkLocation
		(*t.FileMetadata[args.FileName].ChunkHandles)[chunkId] = chunkHandle

		reply.ChunkMap[chunkId] = &common.ClientChunkInfo{
			Location:    chunkLocation,
			ChunkHandle: chunkHandle,
		}
	}

	log.Printf(`Finished saving file/chunk data at master.
	FileMetadata: %v
	ChunkMetadata: %v`,
		t.FileMetadata, t.ChunkMetadata)

	return nil
}

func (t *MasterServer) Delete(args *common.DeleteFileArgsMaster, reply *common.DeleteFileReplyMaster) error {
	return nil
}

func (t *MasterServer) Initialize() {
	t.ChunkServers = map[uint8]*ChunkServerMetadata{}
	t.FileMetadata = map[string]*FileMetadata{}
	t.ChunkMetadata = map[common.ChunkHandle]*ChunkMetadata{}
}

func (t *MasterServer) removeExpiredChunkServers() {
	expiration := time.Now().Unix() - common.ChunkServerExpirationTimeSeconds
	// scan and remove expired chunkServers
	for key, val := range t.ChunkServers {
		// remove expired chunkServers
		if val.TimeStampLastPing < expiration {
			log.Printf("ChunkServer ID %v has expired with LastPingTS %v, currentTS %v", key, val.TimeStampLastPing, expiration)
			delete(t.ChunkServers, key)
		}
	}
}

func mapChunkIdToChunkServerIndex(chunkId uint8, chunkServerIds []uint8) uint8 {
	index := chunkId % uint8(len(chunkServerIds))
	return chunkServerIds[index]
}

func getChunkHandle(chunkName string) common.ChunkHandle {
	h := fnv.New64a()
	h.Write([]byte(chunkName))
	return common.ChunkHandle(h.Sum64())
}

func InitializeMasterServer() {
	server := new(MasterServer)
	server.Initialize()

	// Start listening all incoming traffic with port
	rpc.Register(server)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)
}
