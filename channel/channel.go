package channel

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"time"

	"github.com/asdine/storm/v3"

	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/Ziwei-Wei/cyber-rhizome-host/keygen"
	msglist "github.com/Ziwei-Wei/cyber-rhizome-host/msglist"
	util "github.com/Ziwei-Wei/cyber-rhizome-host/utility"
)

// ChatChannel is a persistent pubsub chat channel
type ChatChannel struct {
	// c info
	channelName     string
	channelType     string
	peerIDToMsgList map[string]*msglist.PeerMessageList

	// user
	userID string

	// services
	host   host.Host
	pubsub *pubsub.PubSub
	sub    *pubsub.Subscription
	db     storm.Node

	// state and chan
	state        string
	pubsubQueue  chan pubsubMessage
	chatQueue    chan chatMessage
	syncMsgQueue chan syncMessage
	stateQueue   chan syncState

	// on handlers
	onNewMessage func(data interface{})
	onSync       func(data interface{})
	onPeerJoin   func(data interface{})
	onPeerLeave  func(data interface{})
}

// OpenChannel create a new channel
func OpenChannel(
	ctx context.Context,
	channelName string,
	channelType string,
	host host.Host,
	p *pubsub.PubSub,
	db *storm.DB,
) (*ChatChannel, error) {
	var data channelData
	err := db.From(channelName).One("ChannelName", channelName, &data)
	if err != nil {
		log.Printf("error: %v in OpenChannel, channelData not found", err)
		return nil, err
	}
	sub, err := join(p, channelName)
	if err != nil {
		log.Printf("error: %v in OpenChannel, join failed", err)
		return nil, err
	}

	// create msglist for each peer
	peerIDToMsgList := make(map[string]*msglist.PeerMessageList)
	for peer := range data.PeerList {
		var peerData msglist.PeerData
		err := db.From(channelName, peer).One("PeerID", peer, &peerData)
		if err != nil {
			if err.Error() == "not found" {
				msgs, err := msglist.New(peer, "unknown", make([]string, 0), db.From(channelName, peer))
				if err != nil {
					log.Printf("error: %v in OpenChannel, New1 failed", err)
					return nil, err
				}
				peerIDToMsgList[peer] = msgs
			} else {
				log.Printf("error: %v in OpenChannel, One failed", err)
			}
		}
		msgs, err := msglist.New(peerData.PeerID, peerData.PeerName, peerData.P2PAddrs, db.From(channelName, peer))
		if err != nil {
			log.Printf("error: %v in OpenChannel, New2 failed", err)
			return nil, err
		}
		peerIDToMsgList[peer] = msgs
	}

	// create our chatchannel
	channel := ChatChannel{
		channelName:     channelName,
		channelType:     channelType,
		peerIDToMsgList: peerIDToMsgList,
		userID:          host.ID().Pretty(),
		host:            host,
		pubsub:          p,
		sub:             sub,
		db:              db.From(channelName),
		state:           "Join",
		pubsubQueue:     make(chan pubsubMessage),
		chatQueue:       make(chan chatMessage),
		syncMsgQueue:    make(chan syncMessage),
		stateQueue:      make(chan syncState),
		onNewMessage:    func(data interface{}) {},
		onSync:          func(data interface{}) {},
		onPeerJoin:      func(data interface{}) {},
		onPeerLeave:     func(data interface{}) {},
	}

	// connect to know addrs, listen to messages
	channel.connectAllPeers()
	channel.host.SetStreamHandler("", channel.handleSyncStream)
	go channel.listenToPeers(ctx)
	go channel.syncChannel(ctx)

	channel.state = "Alive"

	// signal join
	return &channel, nil
}

// CreateChannel create a new c to use
func CreateChannel(
	channelName string,
	channelType string,
	userName string,
	protector string,
	db *storm.DB,
) error {
	PrivKey, err := keygen.CreatePrivKey(userName, protector)
	peerID, err := peer.IDFromPrivateKey(PrivKey)
	channelData := channelData{
		ChannelName: channelName,
		ChannelType: channelType,
		PeerList:    map[string]struct{}{},
	}
	log.Println(channelData)
	msglist.New(
		peerID.Pretty(),
		userName,
		util.BuildP2PAddrsFromString(peerID.Pretty(), []string{"/ip4/127.0.0.1/tcp/6666"}),
		db.From(channelName, peerID.Pretty()),
	)

	// save to db
	err = db.From(channelName).Save(&channelData)
	if err != nil {
		log.Printf("error: %v in CreateChannel, Save failed", err)
		return err
	}
	return db.Close()
}

// Send will send message to all members of the channel
func (c *ChatChannel) Send(content interface{}) error {
	if reflect.TypeOf(content).Kind() == reflect.String {
		tx, err := c.db.From(c.channelName).Begin(true)
		if err != nil {
			log.Printf("error: %v in Send, transaction failed", err)
			return err
		}
		defer tx.Rollback()
		now := time.Now().Unix()
		list := c.peerIDToMsgList[c.userID]
		newMessage := msglist.Message{
			ID:        list.LatestMsgID + 1,
			SenderID:  c.userID,
			AuthorID:  c.userID,
			Content:   content.(string),
			CreatedAt: now,
		}
		list.Add(newMessage)
		list.Save()
		rawMessage := chatRawMessage{
			ID:        newMessage.ID,
			Content:   content.(string),
			CreatedAt: now,
		}

		tx.Commit()
		if err != nil {
			log.Printf("error: %v in Send, transaction failed", err)
			return err
		}

		jsonMsg, err := json.Marshal(&rawMessage)
		if err != nil {
			log.Printf("error: %v in Send, Marshal(&message) failed", err)
			return err
		}

		err = c.pubsub.Publish(c.channelName, jsonMsg)
		if err != nil {
			log.Printf("error: %v in Send, Publish(channel.channelName, jsonChat) failed", err)
			return err
		}
		return nil
	}
	return errors.New("wrong type")
}

// Leave will leave the current channel(can not receive message anymore)
func (c *ChatChannel) Leave() {
	c.state = "Leave"
	c.sendStateInfo()
	c.sub.Cancel()
	c.db.Commit()
}

// On will assign handlers to NewMessage, PeerJoin, PeerLeave
func (c *ChatChannel) On(event string, handler func(data interface{})) error {
	switch event {
	case "NewMessage":
		c.onNewMessage = handler
		break
	case "Sync":
		c.onSync = handler
		break
	case "PeerJoin":
		c.onPeerJoin = handler
		break
	case "PeerLeave":
		c.onPeerJoin = handler
		break
	default:
		return errors.New("event is not supported")
	}
	return nil
}

// AddPeer add a peer to the channel
func (c *ChatChannel) AddPeer(peerID string, peerName string, p2pAddrs []string) error {
	list, err := msglist.New(peerID, peerName, p2pAddrs, c.db.From(peerID))
	if err != nil {
		log.Printf("error: %v in AddPeer, New failed", err)
		return err
	}
	c.peerIDToMsgList[peerID] = list
	peerList := map[string]struct{}{}
	for peer := range c.peerIDToMsgList {
		peerList[peer] = struct{}{}
	}
	err = c.db.Update(&channelData{
		ChannelName: c.channelName,
		ChannelType: c.channelType,
		PeerList:    peerList,
	})
	if err != nil {
		log.Printf("error: %v in AddPeer, Update failed", err)
		return err
	}

	c.connectPeer(util.StringsToMultiAddrs(p2pAddrs))
	return nil
}

// GetPeers get current peers peer ids in the channel
func (c *ChatChannel) GetPeers() []string {
	var ids []string
	for id := range c.peerIDToMsgList {
		ids = append(ids, id)
	}
	return ids
}
