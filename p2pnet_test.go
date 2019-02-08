package p2pnet_test

import (
	"testing"
	"time"

	"github.com/bnsh12/p2pnet"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

var (
	hostaddr   = "127.0.0.1"
	rendezvous = "p2pTest"
)

func initNetwork(host string, port int, rendezvous string) (*p2pnet.Network, error) {
	cfg := &p2pnet.Config{
		ListenHost:       host,
		ListenPort:       port,
		RendezvousString: rendezvous,
	}

	net, err := p2pnet.NewNetwork(cfg)

	return net, err
}

func TestMDNS(t *testing.T) {

	n1, err := initNetwork(hostaddr, 4001, rendezvous)
	if err != nil {
		t.Errorf("Could not initialize network 1")
	}
	n1.InitMDNS()

	n2, err := initNetwork(hostaddr, 4002, rendezvous)
	if err != nil {
		t.Errorf("Could not initialize network 2")
	}

	n2.InitMDNS()

	time.Sleep(time.Millisecond * 500)

	//Check whether n2 found n1 via MDNS
	peerFound := false
	for _, id := range n2.Host.Peerstore().PeersWithAddrs() {

		if id == n1.Host.ID() {
			peerFound = true
		}
	}

	if !peerFound {
		t.Errorf("Could not find peer n1 via MDNS")
	}

	n1.Close()
	n2.Close()
}

func TestDHT(t *testing.T) {

	serName := "testservice"
	// Create a boot node
	bootnode, err := initNetwork(hostaddr, 4003, rendezvous)
	if err != nil {
		t.Errorf("Could not initialize bootnode")
	}

	//create a service provider
	provider, err := initNetwork(hostaddr, 4004, rendezvous)
	if err != nil {
		t.Errorf("Could not initialize provider")
	}

	bNodeInfo := pstore.PeerInfo{
		ID:    bootnode.Host.ID(),
		Addrs: bootnode.Host.Addrs(),
	}

	err = provider.Connect(provider.Context(), bNodeInfo)
	if err != nil {
		t.Errorf("Provider could not connect to bootstrap(%s)", err)
	}

	provider.Router.Advertise(provider.Context(), serName)

	user, err := initNetwork(hostaddr, 4005, rendezvous)
	if err != nil {
		t.Errorf("Could not initialize user")
	}

	err = user.Connect(user.Context(), bNodeInfo)
	if err != nil {
		t.Errorf("Provider could not connect to bootstrap(%s)", err)
	}

	pchan, err := user.Router.FindPeers(user.Context(), serName)
	if err != nil {
		t.Errorf("Could not find peers(%s)", err)
	}

	peer := <-pchan

	if peer.ID != provider.Host.ID() {
		t.Errorf("could not find provider")
	}

	bootnode.Close()
	provider.Close()
	user.Close()
}
