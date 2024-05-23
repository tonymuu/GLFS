package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"glfs/client"
	"glfs/common"
	"io"
	"log"
	"os"
	"strings"
)

// There must be a running GFS cluster
func main() {
	// mode := flag.String("-m", "i", "running mode, i for interactive.")
	// flag.Parse()

	// // only support interactive for now
	// if *mode != "i" {
	// 	os.Exit(1)
	// }

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
		case "create":
			client.Create(inputPath)
		case "delete":
			client.Delete("test_0.dat")
		case "read":
			client.Read("test_0.dat", outputPath)
			checkSum(inputPath, outputPath)
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
