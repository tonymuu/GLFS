package chunkserver

import (
	"glfs/common"
	"log"
)

func (t *ChunkServer) PingMaster() {
	// ping master and join cluster
	log.Print("Started pinging master...")
	// Synchronous call
	args := &common.PingArgs{
		Id:      t.Id,
		Address: t.Address,
	}
	var reply bool
	err := common.DialAndCall("MasterServer.Ping", common.GetMasterServerAddress(), args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Got PingMaster reply ", reply)
}
