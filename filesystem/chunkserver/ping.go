package chunkserver

import (
	"glfs/common"
	"log"
	"net/rpc"
)

func (t *ChunkServer) PingMaster() {
	// ping master and join cluster
	log.Print("Started pinging master...")
	masterClient, err := rpc.DialHTTP("tcp", common.GetMasterServerAddress())
	if err != nil {
		log.Fatal("failed to connect to master:", err)
	}
	// Synchronous call
	args := &common.PingArgs{
		Id:      t.Id,
		Address: t.Address,
	}
	var reply bool
	err = masterClient.Call("MasterServer.Ping", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Got PingMaster reply ", reply)
}
