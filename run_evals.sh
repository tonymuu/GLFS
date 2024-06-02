#!/bin/bash

# parse eval configurations from cmd
scenario=$1
clientcount=$2
iterations=$3
chunkservercount=$4

# # clean up logs
# rm -r ./eval

# # create dir for logs
# mkdir ./eval

echo "All eval are running with $clientcount clients each with $iterations reads"
echo "Starting running evaluations for Scenario: $scenario..."
echo "Setting up cluster with $chunkservercount chunk servers"

# setup cluster
/bin/bash setup_cluster.sh $chunkservercount
# run evals
./build/app -mode e -scenario $scenario -clientcount $clientcount -iterations $iterations
# terminate cluster
echo "Cleaning up"
/bin/bash terminate.sh
echo "Finished running scenario."
