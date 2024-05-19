package main

import (
	"fmt"
	"glfs/common"
	"log"
	"net/rpc"
)

func main() {
	// call master
	masterClient, err := rpc.DialHTTP("tcp", common.GetMasterServerAddress())
	if err != nil {
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	args := &common.PingArgs{}
	var reply bool
	err = masterClient.Call("MasterServer.Ping", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)

	// call chunkserver
	chunkClient, err := rpc.DialHTTP("tcp", common.GetChunkServerAddress(1))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	err = chunkClient.Call("ChunkServer.Ping", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}
