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

// CreateHost create create host and pubsub given privKey
func CreateHost(privKey crypto.PrivKey, prot pnet.Protector, port uint16) (*host.Host, error) {
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

// CreatePubSub create a pubsub for use
func CreatePubSub(host host.Host) (*pubsub.PubSub, error) {
	localPubSub, err := pubsub.NewGossipSub(context.Background(), host)
	if err != nil {
		return nil, err
	}

	return localPubSub, nil
}

/* connection functions */

// ConnectByMultiAddrString create connection to a host
func ConnectByMultiAddrString(h *host.Host, addr string) error {
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

// ConnectByMultiAddrStrings ...
func ConnectByMultiAddrStrings(h *host.Host, multiAddrs []string) error {
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

// ConnectByMultiAddrs ...
func ConnectByMultiAddrs(h *host.Host, multiAddrs []multiaddr.Multiaddr) error {
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

// ConnectWithPeers create connections to hosts
func ConnectWithPeers(h *host.Host, peers map[string][]string) []string {
	var connections []string
	for peer, addr := range peers {
		err := ConnectByMultiAddrStrings(h, addr)
		if err == nil {
			connections = append(connections, peer)
		}

	}
	return connections
}
