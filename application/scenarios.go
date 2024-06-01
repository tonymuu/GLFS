package main

import (
	"glfs/client"
	"log"
	"sync"
)

func ReadOnly(clientCount int) {
	log.Printf("Started initialzing clients")
	clients := initClients(clientCount)
	log.Printf("Initializing clients completed")

	var wg sync.WaitGroup
	log.Printf("All clients started reading")
	for i := 0; i < clientCount; i++ {
		wg.Add(1)

		// each client synchronously sends 100 read requests
		go func() {
			for j := 0; j < 100; j++ {
				clients[j].Read("test_0.dat", ".")
				wg.Done()
			}
		}()
	}
	wg.Wait()

	log.Printf("All clients finished reading")
}

func initClients(clientCount int) []*client.GLFSClient {
	clients := make([]*client.GLFSClient, clientCount)

	var wg sync.WaitGroup
	for i := 0; i < clientCount; i++ {
		wg.Add(1)

		go func() {
			clients[i] = &client.GLFSClient{}
			clients[i].Initialize()
			wg.Done()
		}()
	}
	wg.Wait()

	return clients
}
