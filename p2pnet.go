package p2pnet

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	multiaddr "github.com/multiformats/go-multiaddr"
)

//Network represents libp2p network layer.
type Network struct {
	ctx           context.Context
	cfg           *Config
	addresses     []multiaddr.Multiaddr
	bootstrapDone chan struct{}

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

func readKey(keyfile string) (crypto.PrivKey, error) {

	raw, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(raw)
}

func writeKey(key crypto.PrivKey, keyfile string) {
	raw, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		log.Printf("Could not write key to file (%s) ", err)
	}

	ioutil.WriteFile(keyfile, raw, 0640)
}

//NewNetwork creates a network handle
func NewNetwork(cfg *Config) (*Network, error) {

	n := &Network{cfg: cfg}

	n.ctx = context.Background()

	r := rand.Reader

	var prvKey crypto.PrivKey

	// Check if we can retrieve key from file
	prvKey, err := readKey(cfg.KeyFile)
	if err != nil {
		// Creates a new RSA key pair for this host.
		prvKey, _, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
		if err != nil {
			return nil, err
		}
		writeKey(prvKey, cfg.KeyFile)
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.ListenHost, cfg.ListenPort))

	// Other options can be added here.
	opts := []libp2p.Option{libp2p.ListenAddrs(sourceMultiAddr), libp2p.Identity(prvKey)}

	if cfg.EnableRelay {
		opts = append(opts, libp2p.EnableRelay(circuit.OptActive, circuit.OptHop, circuit.OptDiscovery))
	}

	// libp2p.New constructs a new libp2p Host.
	n.Host, err = libp2p.New(n.ctx, opts...)

	n.StreamMgr = &StreamMgr{host: n.Host}
	n.bootstrapDone = make(chan struct{})

	if err != nil {
		return nil, err
	}

	n.Router, err = initRouter(n.ctx, n.Host)
	if err != nil {
		log.Println("Discovery object initialization failed")
		return nil, err
	}

	for _, addr := range cfg.BootstrapPeers {
		peerinfo, _ := pstore.InfoFromP2pAddr(addr)
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
					close(n.bootstrapDone)
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

//WaitForBootstrap blocks untill local peer is bootstrapped
func (net *Network) WaitForBootstrap() {
	select {
	case <-net.bootstrapDone:
		return
	}
}
