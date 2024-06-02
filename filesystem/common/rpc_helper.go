package common

import (
	"net/rpc"

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

			err = client.Call(rpcName, args, reply)
			return err
		},
	)
}
