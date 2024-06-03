#!/bin/bash

# parse eval configurations from cmd
scenario=$1
clientcount=$2
iterations=$3
chunkservercount=$4
masteravailability=$5
filename=$6

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

echo "Start running clients"
# run evals
./build/app -mode e -scenario $scenario -clientcount $clientcount -iterations $iterations -availability $masteravailability -filename $filename
echo "Finished running clients"

# terminate cluster
echo "Cleaning up"
/bin/bash terminate.sh
echo "Finished running scenario."

printf '=%.0s' {1..100}
printf '\n'


# script to generate file of size for testing
# dd if=/dev/zero of=tmp/app/test_2.dat  bs=24M  count=1
