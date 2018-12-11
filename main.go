package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/multiformats/go-multiaddr"
)

type discoveryNotifee struct {
}

func (n *discoveryNotifee) HandlePeerFound(pi pstore.PeerInfo) {
	fmt.Println("Found peer ", pi.ID.Pretty())
}

func main() {
	help := flag.Bool("help", false, "Display Help")
	listenHost := flag.String("host", "0.0.0.0", "The bootstrap node host listen address\n")
	port := flag.Int("port", 4001, "The bootstrap node listen port")
	flag.Parse()

	if *help {
		fmt.Printf("This is a simple bootstrap node for kad-dht application using libp2p\n\n")
		fmt.Printf("Usage: \n   Run './bootnode'\nor Run './bootnode -host [host] -port [port]'\n")

		os.Exit(0)
	}

	fmt.Printf("[*] Listening on: %s with port: %d\n", *listenHost, *port)

	ctx := context.Background()
	// r := mrand.New(mrand.NewSource(int64(*port))) //Predictive ID
	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", *listenHost, *port))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		panic(err)
	}
	// DHT ..
	// _, err = dht.New(ctx, host)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("")
	fmt.Printf("[*] Your Bootstrap ID Is: /ip4/%s/tcp/%v/p2p/%s\n", *listenHost, *port, host.ID().Pretty())
	// fmt.Println("")

	//mDNS
	ser, err := discovery.NewMdnsService(ctx, host, time.Second*5, "ubox")
	if err != nil {
		panic(err)
	}

	n := &discoveryNotifee{}

	ser.RegisterNotifee(n)

	select {}
}
