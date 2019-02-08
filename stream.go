package p2pnet

import (
	"context"
	"io"
	"log"

	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

//StreamMgr handles connections of network
type StreamMgr struct {
	host host.Host
}

// Stream represents a bidirectional channel between two agents
type Stream interface {
	io.ReadWriter

	// Close closes the stream for writing. Reading will still work (that
	// is, the remote side can still write).
	io.Closer

	// Reset closes both ends of the stream. Use this to tell the remote
	// side to hang up and go away.
	Reset() error
}

// StreamHandler is the type of function used to listen for
// streams opened by the remote side.
type StreamHandler func(Stream)

// SetHandler sets the protocol handler on host.
func (mgr *StreamMgr) SetHandler(pid protocol.ID, handler StreamHandler) {
	mgr.host.SetStreamHandler(pid, func(s inet.Stream) {
		log.Printf("Stream opened for protocol %s", pid)
		handler(s)
	})
}

// RemoveHandler removes the protocol handler
func (mgr *StreamMgr) RemoveHandler(pid protocol.ID) {
	mgr.host.RemoveStreamHandler(pid)
}

// NewStream opens a new stream to given peer
func (mgr *StreamMgr) NewStream(ctx context.Context, peer peer.ID, pid protocol.ID) (Stream, error) {
	return mgr.host.NewStream(ctx, peer, pid)
}
