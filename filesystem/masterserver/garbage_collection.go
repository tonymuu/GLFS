package masterserver

import (
	"glfs/common"
	"glfs/protobufs/pb"
	"log"
	"net/rpc"
	"strings"
	"time"
)

func (t *MasterServer) CleanupDeletedFiles() {
	log.Print("Started cleaning up deleted files...")
	for fileName, file := range t.State.FileMetadata {
		// file not marked for deletion
		if !strings.HasPrefix(file.FileName, ".") {
			continue
		}
		// file marked for deletion, but within retention period
		if file.DeletionTimeStamp > time.Now().Unix()+common.FileRetentionPeriod {
			continue
		}

		log.Printf("Found deleted file %v marked for deletion at %v. Deleting chunks...", file.FileName, file.DeletionTimeStamp)

		t.deleteChunks(file)

		log.Printf("Deleted all chunks for %v. Now cleaning up state", file.FileName)

		delete(t.State.FileMetadata, fileName)

	}
}

func (t *MasterServer) deleteChunks(file *pb.File) {
	for _, chunkHandle := range file.ChunkHandles {
		log.Printf("Deleting chunk handle %v", chunkHandle)
		chunk := t.State.ChunkMetadata[chunkHandle]
		go t.deleteChunk(chunkHandle, chunk)
	}
}

func (t *MasterServer) deleteChunk(chunkHandle uint64, chunk *pb.Chunk) {
	for _, sid := range chunk.ReplicaServerIds {
		chunkServer := t.State.ChunkServers[sid]

		// connect to master server using tcp
		chunkClient, err := rpc.DialHTTP("tcp", chunkServer.ServerAddress)
		if err != nil {
			log.Fatal("Connecting to chunk client failed: ", err)
		}

		args := &common.DeleteFileArgsChunk{
			ChunkHandle: chunkHandle,
		}
		var reply bool

		log.Printf("Deleting file from chunkServer at %v", chunkServer.ServerAddress)
		err = chunkClient.Call("ChunkServer.Delete", args, &reply)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Received DeleteFileReply from chunkServer at %v: %v", chunkServer.ServerAddress, reply)
	}

	log.Printf("Deleted all chunk replicas for %v. Now purging chunk data from master state", chunkHandle)

	// delete chunk data from master state
	delete(t.State.ChunkMetadata, chunkHandle)
}
