package main

import (
	"fmt"
	"glfs/client"
	"glfs/common"
	"sync"
	"time"
)

func ReadOnly(clients []*client.GLFSClient, iterations int, filename string) int64 {
	// upload a test file with one client
	inputPath := common.GetTmpPath("app", filename)
	outputPath := common.GetTmpPath("app", fmt.Sprintf("%v.out", filename))

	clients[0].Create(inputPath)

	// start timer to eval read only scenario
	startTime := time.Now().UnixMilli()

	var wg sync.WaitGroup
	for i := 0; i < len(clients); i++ {
		wg.Add(1)

		// each client synchronously sends 100 read requests
		go func() {
			for j := 0; j < iterations; j++ {
				clients[j].Read(filename, outputPath)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// end timer
	endTime := time.Now().UnixMilli()

	return endTime - startTime
}
