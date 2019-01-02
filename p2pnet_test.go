package p2pnet

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMDNS(t *testing.T) {
	c1 := Config{
		ListenHost:       "127.0.0.1",
		ListenPort:       4001,
		RendezvousString: "p2pTest",
	}

	n1, err := NewNetwork(&c1)
	if err != nil {
		t.Errorf("Could not initialize network 1")
	}
	n1.InitMDNS()

	fmt.Println("Network n1 initialized, ID is", n1.Host().ID().Pretty())

	c2 := c1
	c2.ListenPort++

	n2, err := NewNetwork(&c2)
	if err != nil {
		t.Errorf("Could not initialize network 2")
	}

	n2.InitMDNS()
	fmt.Println("Network n2 initialized, ID is", n2.Host().ID().Pretty())

	time.Sleep(time.Second * 3)

	//Check whether n2 found n1 via MDNS
	peerFound := false
	for _, id := range n2.Host().Peerstore().PeersWithAddrs() {

		if id == n1.Host().ID() {
			peerFound = true
		}
	}

	if !peerFound {
		t.Errorf("Could not find peer n1 via MDNS")
	}
}

func TestMain(m *testing.M) {
	fmt.Println("Start Test")
	os.Exit(m.Run())
}
