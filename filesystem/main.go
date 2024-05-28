package main

import (
	"flag"
	"fmt"
	"glfs/chunkserver"
	"glfs/client"
	"glfs/common"
	"glfs/masterserver"
	"log"
	"os"
)

func main() {
	// read from cmd args
	role := flag.String("role", "", "Can only be master, chunk, or client")
	chunkId := flag.String("id", "", "ChunkServerId")
	flag.Parse()

	// For debugging master
	// *role = "master"

	// For debugging chunk server id 2
	// *role = "chunk"
	// *chunkId = "2"

	// for debugging client
	// *role = "client"

	if len(*role) == 0 {
		log.Fatal("A role must be specified.")
	}
	if *role == "chunk" && len(*chunkId) == 0 {
		log.Fatal("Must provide ID for chunk server")
	}

	// set up logging
	// logDir := fmt.Sprintf("%v/%v/log_%v.txt", common.GetRootDir(), "logs", role)
	name := fmt.Sprintf("%v_%v", *role, *chunkId)
	f, err := os.OpenFile(common.GetLogPath(name), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Printf("main entry")

	log.Printf("Initializing %v", *role)
	switch *role {
	case "master":
		masterserver.InitializeMasterServer()
	case "chunk":
		chunkserver.InitializeChunkServer(chunkId)
	case "client":
		client := client.GLFSClient{}
		client.Initialize()
		client.Write("test_0.dat", 0, []byte{})
	default:
		log.Fatal("role Can only be master, chunk, or client")
	}
}
