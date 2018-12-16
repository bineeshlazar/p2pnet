package main

import (
	"context"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

type discoveryNotifee struct {
	PeerChan chan pstore.PeerInfo
}

func (n *discoveryNotifee) HandlePeerFound(pi pstore.PeerInfo) {
	n.PeerChan <- pi
}

func initMDNS(ctx context.Context, peerhost host.Host, rendezvous string) (<-chan pstore.PeerInfo, error) {
	ser, err := discovery.NewMdnsService(ctx, peerhost, time.Hour, rendezvous)
	if err != nil {
		return nil, err
	}

	n := &discoveryNotifee{}
	n.PeerChan = make(chan pstore.PeerInfo)

	ser.RegisterNotifee(n)
	return n.PeerChan, nil
}
