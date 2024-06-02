#!/bin/bash

# parse eval configurations from cmd
scenario=$1
clientcount=$2
iterations=$3
chunkservercount=$4
masteravailability=$5

# # clean up logs
# rm -r ./eval

# # create dir for logs
# mkdir ./eval

echo "All eval are running with $clientcount clients each with $iterations reads"
echo "Starting running evaluations for Scenario: $scenario, with master availability % $masteravailability..."
echo "Setting up cluster with $chunkservercount chunk servers"

# start up master failure simulation
/bin/bash fail_master.sh $masteravailability &

# setup cluster
/bin/bash setup_cluster.sh $chunkservercount

# run evals
./build/app -mode e -scenario $scenario -clientcount $clientcount -iterations $iterations -availability $masteravailability

# terminate cluster
echo "Cleaning up"
/bin/bash terminate.sh
echo "Finished running scenario."
