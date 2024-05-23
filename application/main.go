package main

import (
	"bufio"
	"flag"
	"fmt"
	"glfs/client"
	"glfs/common"
	"log"
	"os"
	"strings"
)

// There must be a running GFS cluster
func main() {
	mode := flag.String("-m", "i", "running mode, i for interactive.")
	flag.Parse()

	client := client.GLFSClient{}
	client.Initialize()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Enter command, create/delete/read etc. or q to quit.")
		command, _ := reader.ReadString('\n')
		command = strings.Trim(command, "\n")
		fmt.Println("cmd: %v, %v", command, command == "q")

		switch command {
		case "create":
			client.Create(common.GetTmpPath("client", "test_0.dat"))
		case "delete":
			client.Delete("test_0.dat")
		case "read":
			client.Read("test_0.dat")

		case "q":
			os.Exit(0)

		default:
			log.Printf("Testcase not supported, please check spelling!")
		}
	}

}
