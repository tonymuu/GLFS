package common

// type ChunkHandle uint64

// Arg types

type PingArgs struct {
	Id      uint32
	Address string
}

// Arguments for client to master create file call
type CreateFileArgsMaster struct {
	FileName       string
	NumberOfChunks uint32
}

// Arguments for client to master delete file call
type DeleteFileArgsMaster struct {
	FileName string
}

type ReadFileArgsMaster struct {
	FileName string
}

type ReadFileArgsChunk struct {
	ChunkHandle uint64
}

type CreateFileArgsChunk struct {
	ChunkHandle uint64
	Content     []byte
}

type DeleteFileArgsChunk struct {
}

type GetPrimaryArgsMaster struct {
	FileName   string
	ChunkIndex uint64
}

type WriteArgsChunk struct {
	ChunkHandle uint64
	Offset      uint64
	Data        []byte
}

type CommitWriteArgsChunk struct {
	IsPrimary bool
	UpdateId  uint64            // only used when IsPrimary: primary's UpdateId
	Replicas  map[string]uint64 // map from updateId to address
}

// Response types

// Reply from client to master create file call
type CreateFileReplyMaster struct {
	ChunkMap map[uint32]*ClientChunkInfo
}

type DeleteFileReplyMaster struct {
}

type ReadFileReplyMaster struct {
	Chunks []ClientChunkInfo
}

type ReadFileReplyChunk struct {
	Content []byte
}

type WriteReplyChunk struct {
	UpdateId uint64
}

// Objects
type ClientChunkInfo struct {
	ChunkHandle      uint64
	PrimaryLocation  string
	ReplicaLocations []string
}
