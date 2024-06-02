package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"glfs/client"
	"glfs/common"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// There must be a running GFS cluster
func main() {
	mode := flag.String("mode", "", "running mode, i for interactive.")
	scenario := flag.String("scenario", "", "Which evaluation scenarios to run")
	clientCount := flag.String("clientcount", "", "Which evaluation scenarios to run")
	iterations := flag.String("iterations", "", "How many iterations per client?")
	masterAvailability := flag.String("availability", "", "how often does master fail?")
	filename := flag.String("filename", "", "which file to test on?")

	flag.Parse()

	// we can run the app in interactive mode with a single client for debugging
	if *mode == "i" {
		runInteractiveApp()
		return
	}

	// we can also run the app with predefined scenarios for perf evaluations.
	if *mode == "e" {
		f, err := os.OpenFile(fmt.Sprintf("%v/logs/%v", common.GetRootDir(), *scenario), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)

		runEvaluations(*scenario, *clientCount, *iterations, *masterAvailability, *filename)
		return
	}
}

func runEvaluations(scenario string, clientCount string, iterations string, masterAvailability string, filename string) {
	// init clients
	count, _ := strconv.Atoi(clientCount)
	it, _ := strconv.Atoi(iterations)

	log.Printf("Started initialzing clients")
	clients := initClients(count)
	log.Printf("Initializing clients completed")

	var duration int64
	switch scenario {
	case "readonly":
		duration = ReadOnly(clients, it, filename)
	}

	outputStr := fmt.Sprintf("Scenario:%v, clientCount:%v, iterations:%v, duration:%v, masterAvailability:%v, filename:%v",
		scenario, clientCount, it, duration, masterAvailability, filename)
	outputEvalResult(outputStr)
}

func runInteractiveApp() {
	client := client.GLFSClient{}
	client.Initialize()

	inputPath := common.GetTmpPath("app", "test_0.dat")
	outputPath := common.GetTmpPath("app", "test_0.dat.out")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Enter command, create/delete/read etc. or q to quit.")
		command, _ := reader.ReadString('\n')
		command = strings.Trim(command, "\n")

		switch command {
		case "c":
			client.Create(inputPath)
		case "d":
			client.Delete("test_0.dat")
		case "r":
			client.Read("test_0.dat", outputPath)
			checkSum(inputPath, outputPath)
		case "w":
			client.Write("test_0.dat", 0, []byte{})
		case "q":
			os.Exit(0)

		default:
			log.Printf("Testcase not supported, please check spelling!")
		}
	}
}

func checkSum(filePath1 string, filePath2 string) bool {
	sum1, err := md5sum(filePath1)
	if err != nil {
		log.Fatal(err)
	}
	sum2, err := md5sum(filePath2)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("path1 sum %v, path2 sum %v, equal: %v", sum1, sum2, sum1 == sum2)
	return sum1 == sum2
}

func md5sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func outputEvalResult(res string) {
	f, err := os.OpenFile(fmt.Sprintf("%v/eval/eval.txt", common.GetRootDir()), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	f.WriteString(res + "\n")
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
