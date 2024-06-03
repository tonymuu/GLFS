#!/bin/bash
AVAILABILITY_LOW=0
AVAILABILITY_HIGH=100
AVAILABILITY_STEP=20

CLIENT_COUNT_LOW=1
CLIENT_COUNT_HIGH=4
CLIENT_COUNT_STEP=1

ITERATIONS_LOW=20
ITERATIONS_HIGH=20
ITERATIONS_STEP=24

rm ./eval_output.txt

i=1
for avail in $(seq $AVAILABILITY_HIGH -$AVAILABILITY_STEP $AVAILABILITY_LOW); do
    for clientcount in $(seq $CLIENT_COUNT_LOW $CLIENT_COUNT_STEP $CLIENT_COUNT_HIGH); do
        for it in $(seq $ITERATIONS_LOW $ITERATIONS_STEP $ITERATIONS_HIGH); do
            echo "Running $i evaluation: availability:$avail, clientCount:$clientcount, iterations:$it"
            /bin/bash run_evals.sh readonly $clientcount $it 7 $avail test_1.dat >> ./eval_output.txt
        done
    done
done
