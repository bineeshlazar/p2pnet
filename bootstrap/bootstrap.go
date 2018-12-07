package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("rendezvous")

func main() {
	help := flag.Bool("h", false, "Display Help")
	config, err := parseFlags()
	if err != nil {
		panic(err)
	}

	if *help {
		fmt.Printf("This program demonstrates a simple p2p chat application using libp2p\n\n")
		fmt.Printf("Usage: Run './chat in two different terminals. Let them connect to the bootstrap nodes, announce themselves and connect to the peers\n")

		os.Exit(0)
	}

	ctx := context.Background()

	// Configure p2p host
	addrs := make([]multiaddr.Multiaddr, len(config.ListenAddresses))
	copy(addrs, config.ListenAddresses)

	options := []libp2p.Option{libp2p.ListenAddrs(addrs...)}

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(ctx, options...)
	if err != nil {
		panic(err)
	}
	fmt.Println("Host created. We are:", host.ID().Pretty())

	// Start a DHT, for use in peer discovery. We can't just make a new DHT client
	// because we want each peer to maintain its own local copy of the DHT, so
	// that the bootstrapping node of the DHT can go down without inhibitting
	// future peer discovery.
	kademliaDHT, err := libp2pdht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	fmt.Println("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	select {}
}
