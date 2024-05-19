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

func (t *MasterServer) HealthCheck(args *common.HealthCheckArgs, reply *bool) error {
	*reply = true
	return nil
}

func main() {
	server := new(MasterServer)
	rpc.Register(server)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)
}
