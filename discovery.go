package p2pnet

import (
	"context"
	"log"

	p2p "github.com/libp2p/go-libp2p-discovery"
	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

//Discovery handles service advertising and discovery.
type Discovery struct {
	ctx          context.Context
	dht          *libp2pdht.IpfsDHT
	dhtDiscovery *p2p.RoutingDiscovery
}

func initDiscovery(n *Network) (*Discovery, error) {

	d := &Discovery{}
	var err error

	d.ctx = n.ctx

	//Create dht and discovery handles
	d.dht, err = libp2pdht.New(n.ctx, n.host)
	if err != nil {
		log.Println("DHT initialization failed")
		return nil, err
	}
	d.dhtDiscovery = p2p.NewRoutingDiscovery(d.dht)

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	err = d.dht.Bootstrap(n.ctx)
	if err != nil {
		log.Println("DHT bootstrap failed")
		return nil, err
	}

	return d, nil
}

//Advertise a service to DHT
func (d *Discovery) Advertise(service string) {
	p2p.Advertise(d.ctx, d.dhtDiscovery, service)
}

// FindPeers using DHT. Note that channel is not long standing.
// It will get closed after each peer search
func (d *Discovery) FindPeers(service string) (<-chan pstore.PeerInfo, error) {
	return d.dhtDiscovery.FindPeers(d.ctx, service)
}
