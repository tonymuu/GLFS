package main

import (
	"flag"
	"glfs/common"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type ChunkServer struct {
	Id      uint8
	Address string
}

func (t *ChunkServer) Ping(args *common.PingArgs, reply *bool) error {
	*reply = true
	return nil
}

func main() {
	cmd := flag.String("cmd", "", "")
	flag.Parse()
	log.Printf("my cmd: %v\n", string(*cmd))

	// Init chunk server
	chunk := new(ChunkServer)
	id, _ := strconv.Atoi(*cmd)
	chunk.Id = uint8(id)
	chunk.Address = common.GetChunkServerAddress(chunk.Id)

	// ping master and join cluster
	masterClient, err := rpc.DialHTTP("tcp", common.GetMasterServerAddress())
	if err != nil {
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	args := &common.PingArgs{
		Id:      chunk.Id,
		Address: chunk.Address,
	}
	var reply bool
	err = masterClient.Call("MasterServer.Ping", args, &reply)
	if err != nil {
		log.Fatal(err)
	}

	// start serving
	rpc.Register(chunk)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", chunk.Address)
	if err != nil {
		log.Fatal(err)
	}
	http.Serve(l, nil)
}
