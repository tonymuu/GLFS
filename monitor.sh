#!/bin/bash

while true; do
    # find glfs master server acount
    n=$(ps aux | grep "[g]lfs -role master"| wc -l)

    echo "Found $n master server process..."

    # if cannot find master 
    if [ "$n" -lt 1 ]; then
        echo "Restarting master..."
        ./build/glfs -role master &
        echo "Master started."
    fi

    sleep 2
done