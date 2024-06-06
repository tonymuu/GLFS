#!/bin/bash

if [ "$#" -lt 1 ]; then
    echo "You must provide the number of chunk servers as an argument. Something like \"/bin/bash setup_cluster.sh\" 7"
    exit 0
fi

NUM_CHUNK_SERVERS=$1

# clean up logs and tmp folders
rm -r ./logs
rm -r ./tmp/chunk
rm -r ./tmp/master

# create dir for logs
mkdir ./logs

echo "Generating a testing file of 24M at tmp/app/test_2.dat"
mkdir ./tmp/app
dd if=/dev/zero of=tmp/app/test_2.dat  bs=24M  count=1

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
rm ./monitor_output.txt
/bin/bash monitor.sh &>> monitor_output.txt &
echo "Monitor started and outputting logs at .monitor_output.txt."

echo "Cluster is running in background. Run the terminate.sh script to kill all GLFS processes"

# compile proto bufs
# protoc -I=. --go_out=. ./masterserver.proto
