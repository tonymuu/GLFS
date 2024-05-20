package main

import (
	"glfs/common"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type ChunkServer struct{}

func (t *ChunkServer) Ping(args *common.PingArgs, reply *bool) error {
	*reply = true
	return nil
}

func main() {
	// ping master and join cluster
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

	// start serving
	chunk := new(ChunkServer)
	rpc.Register(chunk)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", common.GetChunkServerAddress(1))
	if err != nil {
		log.Fatal(err)
	}
	http.Serve(l, nil)
}
