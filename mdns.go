package main

import (
	"context"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

func initMDNS(ctx context.Context, peerhost host.Host,
	rendezvous string, notifee discovery.Notifee) error {

	ser, err := discovery.NewMdnsService(ctx, peerhost, time.Hour, rendezvous)
	if err == nil {
		ser.RegisterNotifee(notifee)
	}
	return err
}
