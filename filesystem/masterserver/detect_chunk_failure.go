package masterserver

import (
	"glfs/common"
	"glfs/protobufs/pb"
	"log"
	"time"
)

func (t *MasterServer) CleanupFailedChunkServers() {
	log.Print("Started remove failed chunk servers...")
	removeExpiredChunkServers(&t.State.ChunkServers)
	log.Print("Finished remove failed chunk servers")
}

func removeExpiredChunkServers(chunkServers *map[uint32]*pb.ChunkServer) {
	expiration := time.Now().Unix() - common.ChunkServerExpirationTimeSeconds
	// scan and remove expired chunkServers
	for key := range *chunkServers {
		// remove expired chunkServers
		ts := (*chunkServers)[key].TimeStampLastPing
		if ts < expiration {
			log.Printf("ChunkServer ID %v has expired with LastPingTS %v, currentTS %v", key, ts, expiration)
			delete(*chunkServers, key)
		}
	}
}
