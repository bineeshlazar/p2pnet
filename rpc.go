package p2pnet

import (
	"context"
	"fmt"

	gorpc "github.com/libp2p/go-libp2p-gorpc"
	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

/*
RPC represents a RPC object of a host. It represents both server and client
*/
type RPC struct {
	server *gorpc.Server
	client *gorpc.Client
}

func initRPC(host host.Host, rendezvous string) *RPC {

	protoID := protocol.ID(fmt.Sprintf("/p2p/rpc/%s", rendezvous))

	rpc := &RPC{}

	rpc.client = gorpc.NewClient(host, protoID)
	rpc.server = gorpc.NewServer(host, protoID)

	return rpc
}

//Call performs an RPC call to a registered Server service and blocks until completed.
func (rpc *RPC) Call(ctx context.Context, dest peer.ID,
	service, method string,
	args, reply interface{}) error {

	return rpc.client.CallContext(ctx, dest, service, method, args, reply)
}

/*
Register publishes in the server the set of methods of the service interface that satisfy the following conditions:

- exported method of exported type
- two arguments, both of exported type
- the second argument is a pointer
- one return value, of type error
*/
func (rpc *RPC) Register(service interface{}) error {
	return rpc.server.Register(service)
}
