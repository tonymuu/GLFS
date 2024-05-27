package masterserver

import (
	"fmt"
	"glfs/common"
	"glfs/protobufs/pb"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
)

var persistedStateFile *os.File

type MasterServer struct {
	State pb.MasterServer
}

func (t *MasterServer) Ping(args *common.PingArgs, reply *bool) error {
	log.Printf("Master.Ping called with args %v", args)

	t.State.ChunkServers[args.Id] = &pb.ChunkServer{
		ServerId:          args.Id,
		ServerAddress:     args.Address,
		TimeStampLastPing: time.Now().Unix(),
	}

	log.Printf("Updated ChunkServers info. New state: %v", t.State.ChunkServers)

	// Expired chunk servers are presumed dead, so we remove them.
	t.removeExpiredChunkServers()

	*reply = true
	return nil
}

// RPC method on the MasterServer used to create a file.
func (t *MasterServer) Create(args *common.CreateFileArgsMaster, reply *common.CreateFileReplyMaster) error {
	log.Printf("Received Master.Create call with args %v", args)

	// chunkId := uint32(0)
	reply.ChunkMap = map[uint32]*common.ClientChunkInfo{}

	chunkHandles := make([]uint64, args.NumberOfChunks)
	t.State.FileMetadata[args.FileName] = &pb.File{
		FileName:     args.FileName,
		ChunkHandles: chunkHandles,
	}

	chunkServerIds := make([]uint32, len(t.State.ChunkServers))
	i := 0
	for chunkServerId := range t.State.ChunkServers {
		chunkServerIds[i] = chunkServerId
		i++
	}

	for chunkId := range args.NumberOfChunks {
		// compute chunkHandle
		chunkName := fmt.Sprintf("%v-%v", args.FileName, chunkId)
		chunkHandle := getChunkHandle(chunkName)

		// Get chunkServer address
		primaryServerId, replicaServerIds := mapChunkIdToChunkServerIndex(chunkHandle, common.ReplicationGoal, chunkServerIds)
		primaryServer := t.State.ChunkServers[primaryServerId]

		// Save this information
		t.State.ChunkMetadata[chunkHandle] = &pb.Chunk{}
		t.State.ChunkMetadata[chunkHandle].PrimaryServerId = primaryServerId
		t.State.ChunkMetadata[chunkHandle].ReplicaServerIds = make([]uint32, common.ReplicationGoal)

		replicaServerAddresses := make([]string, common.ReplicationGoal)
		for i, sid := range replicaServerIds {
			t.State.ChunkMetadata[chunkHandle].ReplicaServerIds[i] = sid
			replicaServerAddresses[i] = t.State.ChunkServers[sid].ServerAddress
		}

		(t.State.FileMetadata[args.FileName].ChunkHandles)[chunkId] = chunkHandle

		reply.ChunkMap[chunkId] = &common.ClientChunkInfo{
			PrimaryLocation:  primaryServer.ServerAddress,
			ReplicaLocations: replicaServerAddresses,
			ChunkHandle:      chunkHandle,
		}
	}

	if err := t.flushState(); err != nil {
		log.Fatalf("Failed checkpointing master.")
		return err
	}

	log.Printf(`Finished saving file/chunk data at master.
	FileMetadata: %v
	ChunkMetadata: %v`,
		t.State.FileMetadata, t.State.ChunkMetadata)

	return nil
}

// Upon receiving the Delete call, master will only immediately mark the file as to be deleted (with a deletion timestamp).
// Files are retained for 3 days from the second it is marked for deletion (TODO: make this configurable).
// The deletion of the physical copies will be handled by the garbage collection thread and chunkservers.
func (t *MasterServer) Delete(args *common.DeleteFileArgsMaster, reply *bool) error {
	log.Printf("Received Master.Delete call with args %v", args)
	fileInfo, found := t.State.FileMetadata[args.FileName]

	if !found {
		*reply = false
		return fmt.Errorf("file not found with name %v", args.FileName)
	}

	// make the file hidden by adding a period before its name
	fileInfo.FileName = fmt.Sprintf(".%v", fileInfo.FileName)
	// set deletion timestamp
	fileInfo.DeletionTimeStamp = time.Now().Add(time.Hour * 3 * 24).Unix()

	if err := t.flushState(); err != nil {
		log.Fatalf("Failed checkpointing master.")
		return err
	}

	log.Printf(`Finished marking file for deletion at master.
	FileMetadata: %v
	ChunkMetadata: %v`,
		t.State.FileMetadata[args.FileName], t.State.ChunkMetadata)

	*reply = true
	return nil
}

func (t *MasterServer) Read(args *common.ReadFileArgsMaster, reply *common.ReadFileReplyMaster) error {
	log.Printf("Received Master.Delete call with args %v", args)
	fileInfo, found := t.State.FileMetadata[args.FileName]

	if !found {
		return fmt.Errorf("file not found with name %v", args.FileName)
	}

	// TODO: check for chunkserver health here. If primary is not healthy, fall baack to replicas.
	reply.Chunks = make([]common.ClientChunkInfo, len(fileInfo.ChunkHandles))
	for i, chunkHandle := range fileInfo.ChunkHandles {
		chunkServerId := t.State.ChunkMetadata[chunkHandle].PrimaryServerId
		chunkServerAddress := t.State.ChunkServers[chunkServerId].ServerAddress
		reply.Chunks[i] = common.ClientChunkInfo{
			ChunkHandle:     chunkHandle,
			PrimaryLocation: chunkServerAddress,
		}
	}

	log.Printf("Finished reading file at master. Chunks in reply: %v", reply.Chunks)

	return nil
}

func (t *MasterServer) Initialize() {
	t.State.ChunkServers = map[uint32]*pb.ChunkServer{}
	t.State.FileMetadata = map[string]*pb.File{}
	t.State.ChunkMetadata = map[uint64]*pb.Chunk{}

	// On master start, it should check to see if there is any old state.
	stateDir := common.GetTmpPath("master", "")
	err := os.MkdirAll(stateDir, os.ModePerm)
	common.Check(err)

	persistedStateFile, err = os.OpenFile(fmt.Sprintf("%v/state.backup", stateDir), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	common.Check(err)

	t.recoverState()
}

func (t *MasterServer) removeExpiredChunkServers() {
	expiration := time.Now().Unix() - common.ChunkServerExpirationTimeSeconds
	// scan and remove expired chunkServers
	for key, val := range t.State.ChunkServers {
		// remove expired chunkServers
		if val.TimeStampLastPing < expiration {
			log.Printf("ChunkServer ID %v has expired with LastPingTS %v, currentTS %v", key, val.TimeStampLastPing, expiration)
			delete(t.State.ChunkServers, key)
		}
	}
}

// Serialize to protobuf
func (t *MasterServer) flushState() error {
	out, err := proto.Marshal(&t.State)
	if err != nil {
		return err
	}

	if err := os.WriteFile(checkpointPath(), out, 0644); err != nil {
		return err
	}

	log.Printf("Master state checkpointed.")

	return nil
}

// Deserialize from protobuf
func (t *MasterServer) recoverState() error {
	in, err := os.ReadFile(checkpointPath())
	if err != nil {
		return err
	}

	t.State = pb.MasterServer{}
	if err := proto.Unmarshal(in, &t.State); err != nil {
		return err
	}

	log.Printf("Master state recovered: %v", t.State)

	return nil
}

// For now this method hashes chunkHandle into one of the serverids, and place the replicas on servers right after the primary
// TODO: We can scan master's chunkServer state and find the three servers with lowest number of replicas.
// This is okay (close to constant) for small number of chunkservers when n <= 100.
// TODO: use a priorityqueue (min heap) instead for O(logn) lookups when chunk server number is large.
func mapChunkIdToChunkServerIndex(chunkHandle uint64, replicationGoal uint32, chunkServerIds []uint32) (uint32, []uint32) {
	mod := uint64(len(chunkServerIds))
	primaryServerId := chunkServerIds[chunkHandle%mod]
	replicaServerIds := make([]uint32, replicationGoal)
	for i := uint32(0); i < replicationGoal; i++ {
		replicaServerIds[i] = uint32(chunkServerIds[(chunkHandle+uint64(i))%mod])
	}
	return primaryServerId, replicaServerIds
}

func getChunkHandle(chunkName string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(chunkName))
	return uint64(h.Sum64())
}

func checkpointPath() string {
	return common.GetTmpPath("master", "state.checkpoint")
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
