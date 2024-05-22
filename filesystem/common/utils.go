package common

import (
	"fmt"
)

func GetMasterServerAddress() string {
	return MasterServerAddress
}

func GetChunkServerAddress(id uint8) string {
	return fmt.Sprintf(ChunkServerAddress, id)
}

func GetRootDir() string {
	return "/home/mutony/Projects/glfs/filesystem"
}
