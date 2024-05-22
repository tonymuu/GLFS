package main

import (
	"flag"
	"glfs/chunkserver"
	"glfs/client"
	"glfs/masterserver"
	"log"
)

func main() {
	log.Printf("main entry")

	role := flag.String("role", "", "Can only be master, chunk, or client")
	chunkId := flag.String("id", "", "ChunkServerId")
	flag.Parse()
	if len(*role) == 0 {
		log.Fatal("A role must be specified.")
	}
	if *role == "chunk" && len(*chunkId) == 0 {
		log.Fatal("Must provide ID for chunk server")
	}

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
