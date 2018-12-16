package main

import (
	"context"

	discovery "github.com/libp2p/go-libp2p-discovery"
	host "github.com/libp2p/go-libp2p-host"
	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"
)

func initDHT(ctx context.Context, peerhost host.Host, rendezvous string) (*discovery.RoutingDiscovery, error) {
	kademliaDHT, err := libp2pdht.New(ctx, peerhost)
	if err != nil {
		return nil, err
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}

	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	return routingDiscovery, nil
}
