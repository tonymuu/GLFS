package common

// Arg types

type PingArgs struct {
	Id      uint8
	Address string
}

// Arguments for client to master create file call
type CreateFileArgs struct {
	FileName       string
	NumberOfChunks uint8
}

// Arguments for client to master delete file call
type DeleteFileArgs struct {
	FileName string
}

// Arguments for client to master read file call
type ReadFileArgs struct {
}

// Response types

// Reply from client to master create file call
type CreateFileReply struct {
	// Chunkmap maps from a chunkHandle (globally unique 64bit int) to chunkServer address
	ChunkMap map[uint64]string
}

type DeleteFileReply struct {
}

// Objects
