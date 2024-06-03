package common

import (
	"math/rand/v2"
	"net/rpc"
	"time"

	"github.com/avast/retry-go"
)

func DialAndCall(rpcName string, addr string, args any, reply any) error {
	return retry.Do(
		func() error {
			// connect to master server using tcp
			client, err := rpc.DialHTTP("tcp", addr)
			if err != nil {
				return err
			}

			// Artificially create some random "network latency" for better evaluation results
			simualteNetworkLatency()

			err = client.Call(rpcName, args, reply)
			return err
		},
	)
}

func simualteNetworkLatency() {
	// generate a random number between 100-300
	num := 100 + rand.IntN(200)
	// convert the random int into microseconds
	duration := time.Duration(num * int(time.Microsecond))
	time.Sleep(duration)
}
