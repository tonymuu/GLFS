package main

import (
	"flag"
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

	// *role = "master"

	if len(*role) == 0 {
		log.Fatal("A role must be specified.")
	}
	if *role == "chunk" && len(*chunkId) == 0 {
		log.Fatal("Must provide ID for chunk server")
	}

	// set up logging
	// logDir := fmt.Sprintf("%v/%v/log_%v.txt", common.GetRootDir(), "logs", role)
	f, err := os.OpenFile(common.GetLogPath(*role), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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
		client.InitializeClient()
	default:
		log.Fatal("role Can only be master, chunk, or client")
	}
}
