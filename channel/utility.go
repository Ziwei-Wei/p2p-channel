package channel

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	msglist "github.com/cyber-rhizome/p2p-channel/msglist"
	util "github.com/cyber-rhizome/p2p-channel/utility"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

// subscribe a channel as topic using pubsub
func subscribeChannel(p *pubsub.PubSub, channelName string) (*pubsub.Subscription, error) {
	sub, err := p.Subscribe(channelName)
	if err != nil {
		return nil, err
	}
	err = p.RegisterTopicValidator(channelName, func(ctx context.Context, pid peer.ID, msg *pubsub.Message) bool {

		return true
	})
	return sub, nil
}

// get connection with all the peers in the network
func (c *ChatChannel) connectAllPeers(ctx context.Context) {
	for _, list := range c.peerIDToMsgList {
		if c.userID != list.GetPeerID() {
			c.connectPeer(ctx, util.StringsToMultiAddrs(list.P2PAddrs))
		}
	}
}

// ConnectPeer connect with target peer
func (c *ChatChannel) connectPeer(ctx context.Context, multiAddrs []multiaddr.Multiaddr) error {
	addrInfos, err := peer.AddrInfosFromP2pAddrs(multiAddrs...)
	if err != nil {
		log.Printf("error: %v in connectPeer, AddrInfosFromP2pAddrs(multiAddrs...)", err)
		return err
	}

	for _, addrInfo := range addrInfos {
		if addrInfo.ID.Pretty() != c.userID {
			err := c.host.Connect(ctx, addrInfo)
			if err != nil {
				log.Printf("error: %v in connectPeer, Connect", err)
				continue
			}
			return nil
		}
		continue
	}
	return errors.New("connect failed")
}

func print(c *ChatChannel) {
	list := c.peerIDToMsgList[c.userID]
	log.Printf("ChannelName: %s\n", c.channelName)
	log.Printf("UserID: %v\n", c.userID)
	log.Printf("UserP2PAddrs: %v\n", list.P2PAddrs)
	log.Printf("Peers: %v\n", c.peerIDToMsgList)
	log.Printf("PeerP2PAddrs:\n")
	for peer, list := range c.peerIDToMsgList {
		log.Printf("----%v-->%v\n", peer, list.P2PAddrs)
	}
}

func (c *ChatChannel) sendPeerMessage(message pubsubRawMessage) error {
	data, err := json.Marshal(&message)
	if err != nil {
		log.Printf("error: %v at sendPeerMessage, Marshal(&message)", err)
		return err
	}

	err = c.pubsub.Publish(c.channelName, data)
	if err != nil {
		log.Printf("error: %v at sendPeerMessage, Publish(c.cName, data)", err)
		return err
	}
	return nil
}

func (c *ChatChannel) saveChatMessage(msg chatMessage) error {
	return c.peerIDToMsgList[msg.AuthorID].Add(msglist.Message{
		ID:        msg.ID,
		SenderID:  msg.AuthorID,
		AuthorID:  msg.AuthorID,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
	})
}
