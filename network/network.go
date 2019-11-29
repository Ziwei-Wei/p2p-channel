package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	pnet "github.com/libp2p/go-libp2p-core/pnet"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	multiaddr "github.com/multiformats/go-multiaddr"
)

/* network functions */

// createHost create host and pubsub given privKey
func createHost(privKey crypto.PrivKey, prot pnet.Protector, port uint16) (*host.Host, error) {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		libp2p.Identity(privKey),
		libp2p.PrivateNetwork(prot),
		libp2p.EnableAutoRelay(),
	}

	localHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return &localHost, nil
}

// create a pubsub for use
func createPubSub(host host.Host) (*pubsub.PubSub, error) {
	localPubSub, err := pubsub.NewGossipSub(context.Background(), host)
	if err != nil {
		return nil, err
	}

	return localPubSub, nil
}

/* connection functions */

// create connection to a host

func connectByMultiAddrString(h *host.Host, addr string) error {
	multiAddr := multiaddr.StringCast(addr)
	pInfo, err := peerstore.InfoFromP2pAddr(multiAddr)
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
		return err
	}

	err = (*h).Connect(context.Background(), *pInfo)
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
		return err
	}

	return nil
}

func connectByMultiAddrStrings(h *host.Host, multiAddrs []string) error {
	for _, multiAddrString := range multiAddrs {
		multiAddr := multiaddr.StringCast(multiAddrString)
		pInfo, err := peerstore.InfoFromP2pAddr(multiAddr)
		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
			continue
		}

		err = (*h).Connect(context.Background(), *pInfo)
		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
			continue
		}
		return nil
	}
	return errors.New("can not connect to any of addresses")
}

func connectByMultiAddrs(h *host.Host, multiAddrs []multiaddr.Multiaddr) error {
	for _, multiAddr := range multiAddrs {
		pInfo, err := peerstore.InfoFromP2pAddr(multiAddr)
		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
			continue
		}

		err = (*h).Connect(context.Background(), *pInfo)
		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
			continue
		}
		return nil
	}
	return errors.New("can not connect to any of addresses")
}

// create connections to hosts
func connectWithPeers(h *host.Host, peers map[string][]string) []string {
	var connections []string
	for peer, addr := range peers {
		err := connectByMultiAddrStrings(h, addr)
		if err == nil {
			connections = append(connections, peer)
		}

	}
	return connections
}
