package p2pnet

import (
	"context"

	"github.com/multiformats/go-multiaddr"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

//ConnMgr handles connections of network
type ConnMgr struct {
	host host.Host
}

func initConnMgr(ctx context.Context, addr multiaddr.Multiaddr, prvKey crypto.PrivKey) (*ConnMgr, error) {

	conn := &ConnMgr{}
	var err error
	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	conn.host, err = libp2p.New(
		ctx,
		libp2p.ListenAddrs(addr),
		libp2p.Identity(prvKey),
	)

	return conn, err
}

// Connect ensures there is a connection between this host and the peer
func (conn *ConnMgr) Connect(ctx context.Context, pi pstore.PeerInfo) error {
	return conn.host.Connect(ctx, pi)
}
