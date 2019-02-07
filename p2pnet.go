package p2pnet

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

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
	rpc       *RPC
	dht       *Discovery
	conn      *ConnMgr
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

	n.conn, err = initConnMgr(n.ctx, sourceMultiAddr, prvKey)
	if err != nil {
		return nil, err
	}

	n.dht, err = initDiscovery(n.ctx, n.Host())
	if err != nil {
		log.Println("Discovery object initialization failed")
		return nil, err
	}

	if cfg.BootstrapPeer != nil {
		peerinfo, _ := pstore.InfoFromP2pAddr(cfg.BootstrapPeer)
		n.HandlePeerFound(*peerinfo)
	}

	log.Printf("Host ID is %s\n", n.Host().ID().Pretty())

	//Save the our addresses
	addrs := n.Host().Addrs()
	n.addresses = make([]multiaddr.Multiaddr, len(addrs))
	for i, addr := range addrs {
		ipfsAddr, err := multiaddr.NewMultiaddr("/p2p/" + n.Host().ID().Pretty())
		if err != nil {
			panic(err)
		}
		peerAddr := addr.Encapsulate(ipfsAddr)
		n.addresses[i] = peerAddr
	}

	n.rpc = initRPC(n.Host(), cfg.RendezvousString)

	go func() {
		for {
			peerIDs := n.Host().Peerstore().PeersWithAddrs()
			if len(peerIDs) > 0 {

				connected := 0

				for _, id := range peerIDs {
					peer := n.Host().Peerstore().PeerInfo(id)
					err := n.conn.Connect(n.ctx, peer)
					if err == nil {
						addr, _ := pstore.InfoToP2pAddrs(&peer)
						log.Println("Connected to ", addr)
						connected++
					}
				}

				if connected > 0 {
					log.Println("Advertising")
					n.Discovery().Advertise(n.ctx, n.cfg.RendezvousString)
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

	ser, err := discovery.NewMdnsService(net.ctx, net.Host(), time.Hour, net.cfg.RendezvousString)
	if err != nil {
		return err
	}

	ser.RegisterNotifee(net)
	return nil
}

//HandlePeerFound is the Notifee interface for mdns discovery.
//It can be also called to update the peer info found via other ways
func (net *Network) HandlePeerFound(peer pstore.PeerInfo) {

	net.Host().Peerstore().AddAddrs(peer.ID, peer.Addrs, pstore.ProviderAddrTTL)
	log.Println("found", peer.ID)
}

//Addrs returns list of multi addresses we listen on
func (net *Network) Addrs() []multiaddr.Multiaddr {
	return net.addresses
}

//Host returns libp2p host handle. Most applications do not require this
func (net *Network) Host() host.Host {
	return net.conn.host
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
