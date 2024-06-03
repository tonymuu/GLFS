#!/bin/bash

masteravailability=$1

while true; do
    roll=$(($RANDOM % 100))
    echo "Rolled $roll"

    # if roll is less than mastervailability, kill master
    if [ "$masteravailability" -lt "$roll" ]; then
        echo "roll: $roll greater than masteravailability $masteravailability, killing master..."
        kill $(ps aux | grep '[g]lfs -role master' | awk '{print $2}')
    else echo "Roll $roll less than masteravailability $masteravailability, doing nothing";
    fi

    # only executes every 3 seconds
    sleep 3
done
