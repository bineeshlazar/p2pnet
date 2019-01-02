package p2pnet

import (
	"flag"

	multiaddr "github.com/multiformats/go-multiaddr"
)

//Config for the network
type Config struct {
	//RendezvousString is a unique string to identify group of nodes
	RendezvousString string

	//ListenHost is the host IP address network listens on.
	//Use 0.0.0.0 for listening on all interfaces
	ListenHost string

	//ListenPort is the host port to listens on.
	ListenPort int

	//BootstrapPeer is the multiaddress of bootstrap peer if there is any
	BootstrapPeer multiaddr.Multiaddr
}

func parseFlags() *Config {
	c := &Config{}

	var addr string
	flag.StringVar(&c.RendezvousString, "rendezvous", "meetme", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.ListenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&addr, "peer", "", "Adds a peer multiaddress to the bootstrap list")
	flag.IntVar(&c.ListenPort, "port", 4001, "node listen port")

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
