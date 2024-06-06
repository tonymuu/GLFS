package main

import (
	"crypto/rand"
	"fmt"
	"glfs/client"
	"glfs/common"
	math_rand "math/rand/v2"
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
		client := clients[i]
		go read(client, iterations, filename, outputPath, &wg)
	}
	wg.Wait()

	// end timer
	endTime := time.Now().UnixMilli()

	return endTime - startTime
}

func WriteOnly(clients []*client.GLFSClient, iterations int, filename string) int64 {
	// upload a test file with one client
	inputPath := common.GetTmpPath("app", filename)

	clients[0].Create(inputPath)

	// create a 1KB byte array and fill with random bytes
	data := make([]byte, 1024)
	rand.Read(data)

	// start timer to eval read only scenario
	startTime := time.Now().UnixMilli()

	var wg sync.WaitGroup
	for i := 0; i < len(clients); i++ {
		wg.Add(1)
		client := clients[i]
		go write(client, iterations, filename, data, &wg)
	}
	wg.Wait()

	// end timer
	endTime := time.Now().UnixMilli()

	return endTime - startTime
}

// half read, half write
func ReadWrite(clients []*client.GLFSClient, iterations int, filename string) int64 {
	// upload a test file with one client
	inputPath := common.GetTmpPath("app", filename)
	outputPath := common.GetTmpPath("app", fmt.Sprintf("%v.out", filename))

	clients[0].Create(inputPath)

	// create a 1KB byte array and fill with random bytes
	data := make([]byte, 1024)
	rand.Read(data)

	// start timer to eval read only scenario
	startTime := time.Now().UnixMilli()

	var wg sync.WaitGroup
	for i := 0; i < len(clients); i++ {
		wg.Add(1)
		client := clients[i]

		// client at even indices are read clients, odd indices are write clients
		if i%2 == 0 {
			go read(client, iterations, filename, outputPath, &wg)
		} else {
			go write(client, iterations, filename, data, &wg)
		}
	}
	wg.Wait()

	// end timer
	endTime := time.Now().UnixMilli()

	return endTime - startTime
}

func read(client *client.GLFSClient, iterations int, filename string, outputPath string, wg *sync.WaitGroup) {
	for j := 0; j < iterations; j++ {
		client.Read(filename, outputPath)
	}
	wg.Done()
}

func write(client *client.GLFSClient, iterations int, filename string, data []byte, wg *sync.WaitGroup) {
	for j := 0; j < iterations; j++ {
		// generate a random offset for each write
		offset := uint64(math_rand.IntN(common.Test1FileSizeByte / 2))
		client.Write(filename, offset, data)
	}
	wg.Done()
}
