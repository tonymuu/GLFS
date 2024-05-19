package main

import (
	"glfs/common"
	"log"
	"net/rpc"
)

func main() {
	// call master
	masterClient, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	args := &common.HealthCheckArgs{}
	var reply int
	err = masterClient.Call("Arith.Multiply", args, &reply)
	if err != nil {
		log.Fatal(err)
	}

	// call chunkserver
	chunkClient, err := rpc.DialHTTP("tcp", "localhost:1235")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	err = chunkClient.Call("ChunkServer.HealthCheck", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
}
