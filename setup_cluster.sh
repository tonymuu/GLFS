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

# Start n chunk servers use a loop
n=7
for i in `seq 1 $n`
do 
    echo "Starting chunk id $i"
    ./build/glfs -role chunk -id $i &
    echo "chunk id $i started"
done 

# sleep to make sure all servers are ready
sleep 3

# protoc -I=. --go_out=. ./master_checkpoint.proto