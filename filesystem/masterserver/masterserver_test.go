package masterserver

import (
	"glfs/common"
	"glfs/protobufs/pb"
	"sync"
	"testing"
)

func TestInitialize_ShouldInitializeDataStructures(t *testing.T) {
	server := MasterServer{}
	server.Initialize()

	if server.State.ChunkMetadata == nil {
		t.Fatalf("Initialize: must initialize ChunkMetadata")
	}

	if server.State.FileMetadata == nil {
		t.Fatalf("Initialize: must initialize FileMetadata")
	}

	if server.State.ChunkServers == nil {
		t.Fatalf("Initialize: must initialize ChunkServers")
	}
}

func TestPing_ShouldInsertChunkServerMetadata(t *testing.T) {
	server := MasterServer{}
	server.Initialize()

	expectedId, expectedAddress := uint32(8), "chunkServerAddress1"

	var reply bool
	err := server.Ping(&common.PingArgs{
		Id:      expectedId,
		Address: expectedAddress,
	}, &reply)

	if err != nil {
		t.Fatalf("err should be nil but got %v", err)
	}
	if !reply {
		t.Fail()
	}

	val, success := server.State.ChunkServers[expectedId]

	if !success {
		t.Fail()
	}
	if val.ServerAddress != expectedAddress {
		t.Fail()
	}
	if val.TimeStampLastPing == 0 {
		t.Fail()
	}
}

func TestDelete_ShouldChangeHideFile(t *testing.T) {
	server := MasterServer{}
	server.Initialize()

	filename := "some_file"
	server.State.FileMetadata[filename] = &pb.File{
		FileName: filename,
	}

	var reply bool
	err := server.Delete(&common.DeleteFileArgsMaster{
		FileName: filename,
	}, &reply)

	if err != nil {
		t.Fail()
	}
	if server.State.FileMetadata[filename].FileName[0] != '.' {
		t.Fail()
	}
}

func TestDelete_ShouldSetDeletionTimeStamp(t *testing.T) {
	server := MasterServer{}
	server.Initialize()

	filename := "some_file"
	server.State.FileMetadata[filename] = &pb.File{}

	var reply bool
	err := server.Delete(&common.DeleteFileArgsMaster{
		FileName: filename,
	}, &reply)

	if err != nil {
		t.Fail()
	}
	if server.State.FileMetadata[filename].DeletionTimeStamp == 0 {
		t.Fail()
	}
}

func TestDelete_ShouldReturnErrorWhenFileNotFound(t *testing.T) {
	server := MasterServer{}
	server.Initialize()

	var reply bool
	err := server.Delete(&common.DeleteFileArgsMaster{
		FileName: "imaginry_name",
	}, &reply)

	if err == nil || reply {
		t.Fail()
	}
}

func TestDelete_ShouldHandleConcurrentDelete(t *testing.T) {
	server := MasterServer{}
	server.Initialize()

	filename := "some_file"
	server.State.FileMetadata[filename] = &pb.File{
		FileName: "some_file.dat",
	}

	replies := make([]bool, 100)
	var wg sync.WaitGroup

	for i := range replies {
		wg.Add(1)
		go func(rep *bool) {
			server.Delete(&common.DeleteFileArgsMaster{
				FileName: "some_file",
			}, rep)
			defer wg.Done()
		}(&replies[i])
	}

	wg.Wait()

	trueReplyCount := 0
	for _, r := range replies {
		if r {
			trueReplyCount++
		}
	}

	if trueReplyCount != 1 {
		t.Fail()
	}
}
