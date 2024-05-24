package common

import (
	"fmt"
	"log"
)

func GetMasterServerAddress() string {
	return MasterServerAddress
}

func GetChunkServerAddress(id uint32) string {
	return fmt.Sprintf(ChunkServerAddress, id)
}

func GetRootDir() string {
	return "/home/mutony/Projects/glfs"
}

func GetTmpPath(subdir string, filename string) string {
	return fmt.Sprintf("%v/tmp/%v/%v", GetRootDir(), subdir, filename)
}

func GetLogPath(name string) string {
	return fmt.Sprintf("%v/%v/log_%v.txt", GetRootDir(), "logs", name)
}

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
