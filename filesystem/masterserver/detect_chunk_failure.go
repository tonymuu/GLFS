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
		chunkServer := (*chunkServers)[key]
		ts := chunkServer.TimeStampLastPing
		if ts < expiration {
			// Before removing expired server, we should attempt to re-establish connection.
			// Doing this because master might have failed and just recovered, and chunkserver's ping request might have been missed.
			log.Printf("ChunkServer ID %v has expired with LastPingTS %v, currentTS %v. Attemping to reconnect", key, ts, expiration)
			var reply bool
			err := common.DialAndCall("ChunkServer.Ping", chunkServer.ServerAddress, nil, &reply)
			// Re-establish attempt failed, we can safely remove chunkServer.
			if err != nil {
				log.Printf("Failed to reestablish connection with ChunkServer ID %v at %v. Removing from state.", key, chunkServer.ServerAddress)
				delete(*chunkServers, key)
				continue
			}
			log.Printf("Reestablished connection with ChunkServer ID %v at %v. Refreshing timestamp.", key, chunkServer.ServerAddress)
			(*chunkServers)[key].TimeStampLastPing = time.Now().Unix()
		}
	}
}
