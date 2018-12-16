package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	discovery "github.com/libp2p/go-libp2p-discovery"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
)

type network struct {
	ctx          context.Context
	host         host.Host
	cfg          *config
	dhtDiscovery *discovery.RoutingDiscovery
	once         sync.Once
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

	n.dhtDiscovery, err = initDHT(n.ctx, n.host, cfg.RendezvousString)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", cfg.listenHost, cfg.listenPort, n.host.ID().Pretty())

	return n, nil
}

func (net *network) initMDNS() error {
	peerChan, err := initMDNS(net.ctx, net.host, net.cfg.RendezvousString)
	if err != nil {
		return err
	}

	go func() {
		for peer := range peerChan {
			fmt.Print("DNS: ")
			net.updatePeer(peer)
			if err := net.host.Connect(net.ctx, peer); err != nil {
				fmt.Println(err)
			} else {
				net.once.Do(net.pollDHTpeers)
			}
		}
	}()

	return nil
}

func (net *network) advertise(service string) {
	discovery.Advertise(net.ctx, net.dhtDiscovery, service)
}

func (net *network) findPeers(service string) (<-chan pstore.PeerInfo, error) {
	if net.dhtDiscovery == nil {
		return nil, errors.New("Invalid discovery object, DHT initialized?")
	}
	return net.dhtDiscovery.FindPeers(net.ctx, service)
}

func (net *network) updatePeer(peer pstore.PeerInfo) {

	net.host.Peerstore().AddAddrs(peer.ID, peer.Addrs, pstore.ProviderAddrTTL)
	fmt.Println(net.host.Peerstore().PeerInfo(peer.ID))
}

func (net *network) pollDHTpeers() {

	if net.dhtDiscovery == nil {
		fmt.Println("DHT not initialized, skipping DHT polling")
	} else {
		net.advertise(net.cfg.RendezvousString)
		peerChan, err := net.findPeers(net.cfg.RendezvousString)
		if err != nil {
			panic(err)
		}

		go func() {
			for peer := range peerChan {

				if peer.ID == net.host.ID() {
					continue
				}

				fmt.Print("DHT: ")

				net.updatePeer(peer)
			}
		}()
	}
}

func main() {
	help := flag.Bool("help", false, "Display Help")
	cfg := parseFlags()

	if *help {
		fmt.Printf("Simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in local LAN")
		fmt.Printf("Usage: \n   Run './bootnode'\nor Run './bootnode -host [host] -port [port] -peer [multiaddress] -rendezvous [string]'\n")

		os.Exit(0)
	}

	fmt.Printf("[*] Listening on: %s with port: %d\n", cfg.listenHost, cfg.listenPort)

	net, err := newNetwork(cfg)
	if err != nil {
		panic(err)
	}

	if cfg.BootstrapPeer != nil {
		fmt.Println("connecting to ", cfg.BootstrapPeer)
		peerinfo, _ := pstore.InfoFromP2pAddr(cfg.BootstrapPeer)

		if err := net.host.Connect(net.ctx, *peerinfo); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Connection established with bootstrap node:", *peerinfo)
		}
	}

	err = net.initMDNS()
	if err != nil {
		panic(err)
	}

	net.once.Do(net.pollDHTpeers)

	select {}
}
