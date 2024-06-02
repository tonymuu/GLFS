package main

import (
	"glfs/client"
	"glfs/common"
	"sync"
	"time"
)

func ReadOnly(clients []*client.GLFSClient, iterations int) int64 {
	// upload a test file with one client
	inputPath := common.GetTmpPath("app", "test_0.dat")
	outputPath := common.GetTmpPath("app", "test_0.dat.out")

	clients[0].Create(inputPath)

	// start timer to eval read only scenario
	startTime := time.Now().UnixMilli()

	var wg sync.WaitGroup
	for i := 0; i < len(clients); i++ {
		wg.Add(1)

		// each client synchronously sends 100 read requests
		go func() {
			for j := 0; j < iterations; j++ {
				clients[j].Read("test_0.dat", outputPath)
				wg.Done()
			}
		}()
	}
	wg.Wait()

	// end timer
	endTime := time.Now().UnixMilli()

	return endTime - startTime
}
