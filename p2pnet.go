package p2pnet

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	multiaddr "github.com/multiformats/go-multiaddr"
)

//Network represents libp2p network layer.
type Network struct {
	ctx       context.Context
	cfg       *Config
	addresses []multiaddr.Multiaddr

	//ID of this Host
	UUID string

	//Stream manager
	StreamMgr *StreamMgr

	//RPC object of network
	RPC *RPC

	//Router object of network
	Router *Router

	//libp2p host handle. We dont need this mostly
	Host host.Host
}

//NewNetwork creates a network handle
func NewNetwork(cfg *Config) (*Network, error) {

	n := &Network{cfg: cfg}

	n.ctx = context.Background()

	// r := mrand.New(mrand.NewSource(int64(*port))) //Predictive ID
	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.ListenHost, cfg.ListenPort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	n.Host, err = libp2p.New(
		n.ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	n.StreamMgr = &StreamMgr{host: n.Host}

	if err != nil {
		return nil, err
	}

	n.Router, err = initRouter(n.ctx, n.Host)
	if err != nil {
		log.Println("Discovery object initialization failed")
		return nil, err
	}

	if cfg.BootstrapPeer != nil {
		peerinfo, _ := pstore.InfoFromP2pAddr(cfg.BootstrapPeer)
		n.HandlePeerFound(*peerinfo)
	}

	n.UUID = n.Host.ID().Pretty()
	log.Printf("Host ID is %s\n", n.UUID)

	//Save the our addresses
	addrs := n.Host.Addrs()
	n.addresses = make([]multiaddr.Multiaddr, len(addrs))
	for i, addr := range addrs {
		ipfsAddr, err := multiaddr.NewMultiaddr("/p2p/" + n.Host.ID().Pretty())
		if err != nil {
			panic(err)
		}
		peerAddr := addr.Encapsulate(ipfsAddr)
		n.addresses[i] = peerAddr
	}

	n.RPC = initRPC(n.Host, cfg.RendezvousString)

	go func() {
		for {
			peerIDs := n.Host.Peerstore().PeersWithAddrs()
			if len(peerIDs) > 0 {

				connected := 0

				for _, id := range peerIDs {
					peer := n.Host.Peerstore().PeerInfo(id)
					err := n.Connect(n.ctx, peer)
					if err == nil {
						log.Println(n.Host.ID(), " Connected to ", peer.ID)
						connected++
					}
				}

				if connected > 0 {
					log.Println("Advertising")
					n.Router.Advertise(n.ctx, n.cfg.RendezvousString)
					break
				}

			}
			time.Sleep(time.Second * 3)
		}
	}()

	return n, nil
}

//InitMDNS initializes the MDNS discovery in the network
func (net *Network) InitMDNS() error {

	ser, err := discovery.NewMdnsService(net.ctx, net.Host, time.Hour, net.cfg.RendezvousString)
	if err != nil {
		return err
	}

	ser.RegisterNotifee(net)
	return nil
}

//HandlePeerFound is the Notifee interface for mdns discovery.
//It can be also called to update the peer info found via other ways
func (net *Network) HandlePeerFound(peer pstore.PeerInfo) {

	net.Host.Peerstore().AddAddrs(peer.ID, peer.Addrs, pstore.ProviderAddrTTL)
	log.Println("found", peer.ID)
}

//Addrs returns list of multi addresses we listen on
func (net *Network) Addrs() []multiaddr.Multiaddr {
	return net.addresses
}

//Context returns parent context of Network
func (net *Network) Context() context.Context {
	return net.ctx
}

//Connect ensures a connection to remote peer
func (net *Network) Connect(ctx context.Context, peer pstore.PeerInfo) error {
	return net.Host.Connect(ctx, peer)
}

//Close shut down the network and services
func (net *Network) Close() error {
	return net.Host.Close()
}
