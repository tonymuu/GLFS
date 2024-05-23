package common

type ChunkHandle uint64

// Arg types

type PingArgs struct {
	Id      uint8
	Address string
}

// Arguments for client to master create file call
type CreateFileArgsMaster struct {
	FileName       string
	NumberOfChunks uint8
}

// Arguments for client to master delete file call
type DeleteFileArgsMaster struct {
	FileName string
}

type ReadFileArgsMaster struct {
	FileName string
}

type ReadFileArgsChunk struct {
	ChunkHandle
}

type CreateFileArgsChunk struct {
	ChunkHandle ChunkHandle
	Content     []byte
}

type DeleteFileArgsChunk struct {
}

// Arguments for client to master read file call
type ReadFileArgs struct {
}

// Response types

// Reply from client to master create file call
type CreateFileReplyMaster struct {
	ChunkMap map[uint8]*ClientChunkInfo
}

type DeleteFileReplyMaster struct {
}

type ReadFileReplyMaster struct {
	Chunks []ClientChunkInfo
}

type ReadFileReplyChunk struct {
	Content []byte
}

// Objects
type ClientChunkInfo struct {
	ChunkHandle ChunkHandle
	Location    string
}
