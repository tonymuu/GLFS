#!/bin/bash

# clean and build the executable
go clean
go build -C ./application/ -o ../build/app

# clean up logs
rm -r ./eval

# create dir for logs
mkdir ./eval

echo "All eval are running with 10 clients each with 100 reads"
echo "Starting running evaluations..."
echo "Scenario: no master failure"
echo "Setting up cluster with 10 chunk servers"

# setup cluster
/bin/bash setup_cluster.sh 10
# run evals
./build/app -mode e -scenario readonly -clientcount 10
# terminate cluster
echo "Cleaning up"
/bin/bash terminate.sh
echo "Finished running scenario."



echo "Scenario: occasional master failure"

echo "Scenario: frequent master failure"


