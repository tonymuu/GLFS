package masterserver

import "log"

func (t *MasterServer) CleanupFailedChunkServers() {
	log.Print("Started remove failed chunk servers...")
}
