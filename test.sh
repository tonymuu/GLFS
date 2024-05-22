# clean and build the executable
go clean
go build -C ./filesystem/ -o ../build/glfs

# clean up logs
rm -r ./logs

# create dir for logs
mkdir ./logs

echo "Starting master..."
./build/glfs -role master &
echo "master started."

# Sleep to ensure master is started 
sleep 1

echo "Starting chunk id 1"
./build/glfs -role chunk -id 1 &
echo "chunk id 1 started"

echo "Starting chunk id 2"
./build/glfs -role chunk -id 2 &
echo "chunk id 2 started"

# sleep to make sure all servers are ready
sleep 3

echo "Starting client"
./build/glfs -role client -id 2
echo "client started"

# kill $(ps aux | grep [g]o | awk '{print $2}')
# kill $(ps aux | grep [g]lfs | awk '{print $2}')