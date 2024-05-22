package common

import (
	"fmt"
	"log"
)

func GetMasterServerAddress() string {
	return MasterServerAddress
}

func GetChunkServerAddress(id uint8) string {
	return fmt.Sprintf(ChunkServerAddress, id)
}

func GetRootDir() string {
	return "/home/mutony/Projects/glfs/filesystem/tmp"
}

func GetPath(subdir string, filename string) string {
	return fmt.Sprintf("%v/%v/%v", GetRootDir(), subdir, filename)
}

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
