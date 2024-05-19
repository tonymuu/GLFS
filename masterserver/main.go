package main

import (
	"glfs/common"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type MasterServer struct {
}

func (t *MasterServer) Ping(args *common.PingArgs, reply *bool) error {
	*reply = true
	return nil
}

func main() {
	server := new(MasterServer)
	rpc.Register(server)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)
}
