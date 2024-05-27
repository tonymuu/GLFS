package common

// Use 64KB as chunk size for testing
const ChunkSize = 64 * 1024

const ChunkServerExpirationTimeSeconds = 60 * 60 * 7

const MasterServerAddress = "127.0.0.1:1234"
const ChunkServerAddress = "127.0.1.%d:1235"

const ReplicationGoal = 3

// in seconds
const LeaseDuration int64 = 30
