package main

import (
	"flag"
	"glfs/client"
	"glfs/common"
	"log"
)

// There must be a running GFS cluster
func main() {
	testcase := flag.String("t", "", "Which testcase to run?")
	flag.Parse()

	client := client.GLFSClient{}
	client.Initialize()

	switch *testcase {
	case "simple_create":
		client.Create(common.GetTmpPath("client", "test_0.dat"))
	case "simple_delete":
		client.Create(common.GetTmpPath("client", "test_0.dat"))
		client.Delete("test_0.dat")
	default:
		log.Printf("Testcase not supported, please check spelling!")
	}
}
