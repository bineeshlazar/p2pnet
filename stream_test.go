package p2pnet_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/bnsh12/p2pnet"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

func handler(stream p2pnet.Stream) {
	go io.Copy(stream, stream)
}

func TestStream(t *testing.T) {

	server, err := initNetwork(hostaddr, 4001, rendezvous)
	if err != nil {
		t.Errorf("Could not initialize network 1")
	}

	server.StreamMgr.SetHandler("/test/stream", handler)

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

	args := make([]byte, 512)
	reply := make([]byte, 512)

	_, err = rand.Read(args)
	if err != nil {
		t.Errorf("Could not read random data\n%s", err)
	}

	stream, err := client.StreamMgr.NewStream(client.Context(), server.Host.ID(), "/test/stream")
	if err != nil {
		t.Errorf("Could not open stream\n%s", err)
	}

	go stream.Write(args)
	stream.Read(reply)

	if !bytes.Equal(reply, args) {
		t.Error("Received wrong amount of bytes back!")
	}

	server.Close()
	client.Close()
}
