package p2pnet

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	p2p "github.com/libp2p/go-libp2p-discovery"
	host "github.com/libp2p/go-libp2p-host"
	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	multiaddr "github.com/multiformats/go-multiaddr"
)

//Network represents libp2p network layer.
type Network struct {
	ctx          context.Context
	host         host.Host
	cfg          *Config
	dht          *libp2pdht.IpfsDHT
	dhtDiscovery *p2p.RoutingDiscovery
	addresses    []multiaddr.Multiaddr
	rpc          *RPC
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

	//Create dht and discovery handles
	n.dht, err = libp2pdht.New(n.ctx, n.host)
	if err != nil {
		return nil, err
	}
	n.dhtDiscovery = p2p.NewRoutingDiscovery(n.dht)

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	err = n.dht.Bootstrap(n.ctx)
	if err != nil {
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
					n.Advertise(n.cfg.RendezvousString)
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

//Advertise a service to DHT
func (net *Network) Advertise(service string) {
	if net.dhtDiscovery == nil {
		log.Println("DHT not initialized, skipping DHT advertise")
	} else {
		p2p.Advertise(net.ctx, net.dhtDiscovery, service)
	}
}

// FindPeers using DHT. Note that channel is not long standing.
// It will get closed after each peer search
func (net *Network) FindPeers(service string) (<-chan pstore.PeerInfo, error) {
	if net.dhtDiscovery == nil {
		return nil, errors.New("Invalid discovery object, DHT initialized?")
	}
	return net.dhtDiscovery.FindPeers(net.ctx, service)
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

//Host of network
func (net *Network) Host() host.Host {
	return net.host
}

//Discovery return low-level DHT discovery handle of network.
//Most applications may not need this, they can use Advertise and FindPeers methods of Network
func (net *Network) Discovery() *p2p.RoutingDiscovery {
	return net.dhtDiscovery
}

//Context returns parent context of Network
func (net *Network) Context() context.Context {
	return net.ctx
}

//RPC object of network
func (net *Network) RPC() *RPC {
	return net.rpc
}
