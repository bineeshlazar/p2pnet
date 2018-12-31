package p2pnet

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	p2p "github.com/libp2p/go-libp2p-discovery"
	host "github.com/libp2p/go-libp2p-host"
	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/multiformats/go-multiaddr"
)

type network struct {
	ctx          context.Context
	host         host.Host
	cfg          *config
	dht          *libp2pdht.IpfsDHT
	dhtDiscovery *p2p.RoutingDiscovery
}

func newNetwork(cfg *config) (*network, error) {

	n := &network{cfg: cfg}

	n.ctx = context.Background()

	// r := mrand.New(mrand.NewSource(int64(*port))) //Predictive ID
	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.listenHost, cfg.listenPort))

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

	fmt.Printf("[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", cfg.listenHost, cfg.listenPort, n.host.ID().Pretty())

	go func() {
		for {
			time.Sleep(time.Second * 3)
			peerIDs := n.host.Peerstore().PeersWithAddrs()
			if len(peerIDs) > 0 {

				connected := 0

				for _, id := range peerIDs {
					peer := n.host.Peerstore().PeerInfo(id)
					err := n.host.Connect(n.ctx, peer)
					if err == nil {
						addr, _ := pstore.InfoToP2pAddrs(&peer)
						fmt.Println("Connected to ", addr)
						connected++
					}
				}

				if connected > 0 {
					fmt.Println("Advertising")
					n.advertise(n.cfg.RendezvousString)
					break
				}

			}
		}
	}()

	return n, nil
}

func (net *network) initMDNS() error {

	ser, err := discovery.NewMdnsService(net.ctx, net.host, time.Hour, net.cfg.RendezvousString)
	if err != nil {
		return err
	}

	ser.RegisterNotifee(net)
	return nil
}

func (net *network) advertise(service string) {
	if net.dhtDiscovery == nil {
		fmt.Println("DHT not initialized, skipping DHT advertise")
	} else {
		p2p.Advertise(net.ctx, net.dhtDiscovery, service)
	}
}

// Find peers using DHT. Note that channel will get closed after peer search
func (net *network) findPeers(service string) (<-chan pstore.PeerInfo, error) {
	if net.dhtDiscovery == nil {
		return nil, errors.New("Invalid discovery object, DHT initialized?")
	}
	return net.dhtDiscovery.FindPeers(net.ctx, service)
}

func (net *network) HandlePeerFound(peer pstore.PeerInfo) {

	net.host.Peerstore().AddAddrs(peer.ID, peer.Addrs, pstore.ProviderAddrTTL)
	fmt.Println("found", net.host.Peerstore().PeerInfo(peer.ID))
}
