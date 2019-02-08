package p2pnet

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p-host"

	p2p "github.com/libp2p/go-libp2p-discovery"
	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

//Router handles service advertising and discovery.
type Router struct {
	dht          *libp2pdht.IpfsDHT
	dhtDiscovery *p2p.RoutingDiscovery
}

func initRouter(ctx context.Context, host host.Host) (*Router, error) {

	d := &Router{}
	var err error

	//Create dht and discovery handles
	d.dht, err = libp2pdht.New(ctx, host)
	if err != nil {
		log.Println("DHT initialization failed")
		return nil, err
	}
	d.dhtDiscovery = p2p.NewRoutingDiscovery(d.dht)

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	err = d.dht.Bootstrap(ctx)
	if err != nil {
		log.Println("DHT bootstrap failed")
		return nil, err
	}

	return d, nil
}

//Advertise a service to DHT
func (r *Router) Advertise(ctx context.Context, service string) {
	p2p.Advertise(ctx, r.dhtDiscovery, service)
}

// FindPeers using DHT. Note that channel is not long standing.
// It will get closed after each peer search
func (r *Router) FindPeers(ctx context.Context, service string) (<-chan pstore.PeerInfo, error) {
	return r.dhtDiscovery.FindPeers(ctx, service)
}
