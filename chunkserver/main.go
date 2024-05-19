package main

import (
	"glfs/common"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type ChunkServer struct{}

func (t *ChunkServer) HealthCheck(args *common.HealthCheckArgs, reply *bool) error {
	*reply = true
	return nil
}

func main() {
	// ping master and join cluster
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

	// start serving
	chunk := new(ChunkServer)
	rpc.Register(chunk)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":1235")
	if err != nil {
		log.Fatal(err)
	}
	http.Serve(l, nil)
}
