#!/bin/bash

NUM_CHUNK_SERVERS=$1

# clean up logs and tmp folders
rm -r ./logs
rm -r ./tmp/chunk
rm -r ./tmp/master

# create dir for logs
mkdir ./logs

echo "Starting master..."
./build/glfs -role master &
echo "master started."

# Sleep to ensure master is started 
sleep 1

# Start n chunk servers use a loop
for i in `seq 1 $NUM_CHUNK_SERVERS`
do 
    echo "Starting chunk id $i"
    ./build/glfs -role chunk -id $i &
    echo "chunk id $i started"
done 

# sleep to make sure all servers are ready
sleep 3

# Now start up the monitor
/bin/bash monitor.sh &

# compile proto bufs
# protoc -I=. --go_out=. ./masterserver.proto
