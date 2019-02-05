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
	host      host.Host
	cfg       *Config
	addresses []multiaddr.Multiaddr
	rpc       *RPC
	dht       *Discovery
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
	n.host, err = libp2p.New(
		n.ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		return nil, err
	}

	n.dht, err = initDiscovery(n)
	if err != nil {
		log.Println("Discovery object initialization failed")
		return nil, err
	}

	if cfg.BootstrapPeer != nil {
		peerinfo, _ := pstore.InfoFromP2pAddr(cfg.BootstrapPeer)
		n.HandlePeerFound(*peerinfo)
	}

	log.Printf("Host ID is %s\n", n.host.ID().Pretty())

	//Save the our addresses
	addrs := n.host.Addrs()
	n.addresses = make([]multiaddr.Multiaddr, len(addrs))
	for i, addr := range addrs {
		ipfsAddr, err := multiaddr.NewMultiaddr("/p2p/" + n.host.ID().Pretty())
		if err != nil {
			panic(err)
		}
		peerAddr := addr.Encapsulate(ipfsAddr)
		n.addresses[i] = peerAddr
	}

	n.rpc = initRPC(n.host, cfg.RendezvousString)

	go func() {
		for {
			peerIDs := n.host.Peerstore().PeersWithAddrs()
			if len(peerIDs) > 0 {

				connected := 0

				for _, id := range peerIDs {
					peer := n.host.Peerstore().PeerInfo(id)
					err := n.host.Connect(n.ctx, peer)
					if err == nil {
						addr, _ := pstore.InfoToP2pAddrs(&peer)
						log.Println("Connected to ", addr)
						connected++
					}
				}

				if connected > 0 {
					log.Println("Advertising")
					n.Discovery().Advertise(n.cfg.RendezvousString)
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

	ser, err := discovery.NewMdnsService(net.ctx, net.host, time.Hour, net.cfg.RendezvousString)
	if err != nil {
		return err
	}

	ser.RegisterNotifee(net)
	return nil
}

//HandlePeerFound is the Notifee interface for mdns discovery.
//It can be also called to update the peer info found via other ways
func (net *Network) HandlePeerFound(peer pstore.PeerInfo) {

	net.host.Peerstore().AddAddrs(peer.ID, peer.Addrs, pstore.ProviderAddrTTL)
	log.Println("found", net.host.Peerstore().PeerInfo(peer.ID))
}

//Addrs returns list of multi addresses we listen on
func (net *Network) Addrs() []multiaddr.Multiaddr {
	return net.addresses
}

//Host returns libp2p host handle. Most applications do not require this
func (net *Network) Host() host.Host {
	return net.host
}

//Context returns parent context of Network
func (net *Network) Context() context.Context {
	return net.ctx
}

//RPC object of network
func (net *Network) RPC() *RPC {
	return net.rpc
}

//Discovery object of network
func (net *Network) Discovery() *Discovery {
	return net.dht
}
