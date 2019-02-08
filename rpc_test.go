package p2pnet_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"log"
	"testing"

	pstore "github.com/libp2p/go-libp2p-peerstore"
)

type PingArgs struct {
	Data []byte
}
type PingReply struct {
	Data []byte
}
type PingService struct{}

func (t *PingService) Ping(ctx context.Context, argType PingArgs, replyType *PingReply) error {
	log.Println("Received a Ping call")
	replyType.Data = argType.Data
	return nil
}

func TestRPC(t *testing.T) {

	server, err := initNetwork(hostaddr, 4001, rendezvous)
	if err != nil {
		t.Errorf("Could not initialize network 1")
	}

	svc := PingService{}
	err = server.RPC.Register(&svc)
	if err != nil {
		t.Errorf("Could not register service\n%s", err)
	}

	client, err := initNetwork(hostaddr, 4002, rendezvous)
	if err != nil {
		t.Errorf("Could not initialize network 2")
	}

	serverInfo := pstore.PeerInfo{
		ID:    server.Host.ID(),
		Addrs: server.Host.Addrs(),
	}
	err = client.Connect(client.Context(), serverInfo)
	if err != nil {
		t.Errorf("client could not connect to server(%s)", err)
	}

	var reply PingReply
	var args PingArgs

	args.Data = make([]byte, 64)
	_, err = rand.Read(args.Data)
	if err != nil {
		t.Errorf("Could not read random data\n%s", err)
	}

	err = client.RPC.Call(client.Context(), serverInfo.ID, "PingService", "Ping", args, &reply)
	if err != nil {
		t.Errorf("Could not do RPC\n%s", err)
	}

	if !bytes.Equal(reply.Data, args.Data) {
		t.Error("Received wrong amount of bytes back!")
	}

	server.Close()
	client.Close()
}
