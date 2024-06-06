package common

// Use 1MB as chunk size for testing
const ChunkSize = 1 * 1024 * 1024

const ChunkServerExpirationTimeSeconds = 60 * 60 * 7

const MasterServerAddress = "127.0.0.1:1234"
const ChunkServerAddress = "127.0.1.%d:1235"

const ReplicationGoal = 3

// in seconds
const LeaseDuration int64 = 30

// For testing convenience, set it to 5 seconds. Canonically this is 3 days in the GFS paper.
const FileRetentionPeriod = 5

const Test1FileSizeByte = 46 * 1024 * 1014
