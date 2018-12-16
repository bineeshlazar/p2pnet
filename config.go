package main

import (
	"flag"

	"github.com/multiformats/go-multiaddr"
)

type config struct {
	RendezvousString string
	ProtocolID       string
	listenHost       string
	listenPort       int
	BootstrapPeer    multiaddr.Multiaddr
}

func parseFlags() *config {
	c := &config{}

	var addr string
	flag.StringVar(&c.RendezvousString, "rendezvous", "meetme", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.listenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/chat/1.1.0", "Sets a protocol id for stream headers")
	flag.StringVar(&addr, "peer", "", "Adds a peer multiaddress to the bootstrap list")
	flag.IntVar(&c.listenPort, "port", 4001, "node listen port")

	flag.Parse()
	if len(addr) > 0 {
		var err error
		c.BootstrapPeer, err = multiaddr.NewMultiaddr(addr)
		if err != nil {
			panic(err)
		}
	}
	return c
}
