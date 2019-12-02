package channel

import (
	"context"
	"errors"
	"log"

	util "github.com/Ziwei-Wei/cyber-rhizome-host/utility"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

// subscribe a channel as topic using pubsub
func join(p *pubsub.PubSub, channelName string) (*pubsub.Subscription, error) {
	sub, err := p.Subscribe(channelName)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// get connection with all the peers in the network
func (c *ChatChannel) connectAllPeers() {
	for peerID, list := range c.peerIDToMsgList {
		if peerID != list.GetPeerID() {
			c.connectPeer(util.StringsToMultiAddrs(list.P2PAddrs))
		}
	}
}

// ConnectPeer connect with target peer
func (c *ChatChannel) connectPeer(multiAddrs []multiaddr.Multiaddr) error {
	addrInfos, err := peer.AddrInfosFromP2pAddrs(multiAddrs...)
	if err != nil {
		log.Printf("error: %v in connectPeer, AddrInfosFromP2pAddrs(multiAddrs...)", err)
		return err
	}
	for _, addrInfo := range addrInfos {
		err := c.host.Connect(context.Background(), addrInfo)
		if err != nil {
			log.Printf("error: %v in connectPeer, AddrInfosFromP2pAddrs(multiAddrs...)", err)
			continue
		}
		log.Printf("connected to %s", addrInfo.String())
		return nil
	}
	return errors.New("connect failed")
}
