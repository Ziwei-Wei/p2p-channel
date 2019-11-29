package channel

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/asdine/storm/v3"

	host "github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatChannel is a persistent pubsub chat channel
type ChatChannel struct {
	// channel info
	channelName   string
	channelType   string
	memberToPID   map[string]string
	memberToAddrs map[string][]string

	// user
	userName string

	// services
	host   *host.Host
	pubsub *pubsub.PubSub
	sub    *pubsub.Subscription
	db     *storm.DB

	// state
	state    string
	msgQueue []message
	mutex    sync.Mutex
}

// OpenChannel create a new channel
func OpenChannel(
	userName string,
	channelName string,
	channelType string,
	db *storm.DB,
	host *host.Host,
	p *pubsub.PubSub,
) (Channel, error) {
	var info channelInfo
	err := db.One("ChannelName", channelName, &info)
	if err != nil {
		return nil, err
	}

	sub, err := join(p, channelName)
	if err != nil {
		return nil, err
	}
	connect(host, info.MemberToAddrs)

	channel := ChatChannel{
		channelName:   channelName,
		channelType:   channelType,
		userName:      userName,
		host:          host,
		pubsub:        p,
		sub:           sub,
		db:            db,
		memberToPID:   info.MemberToPID,
		memberToAddrs: info.MemberToAddrs,
		state:         "Not Ready",
	}

	go ChatChannel.listenToPeers()

	return &channel, nil
}

// Send will send message to all members of the channel
func (channel *ChatChannel) Send(data []byte) error {
	db, err := channel.db.Begin(true)
	if err != nil {
		return err
	}
	defer db.Rollback()

	var info peerMessageInfo
	err = db.From(channel.channelName).One("PeerName", channel.userName, &info)
	if err != nil {
		log.Println(err)
		return err
	}

	var latestID int = info.RecorderInfo[channel.userName].LatestMsgID + 1
	var createdAt int64 = time.Now().Unix()

	var message message = message{
		ID:        latestID,
		Author:    channel.userName,
		Content:   data,
		CreatedAt: createdAt,
	}
	info.RecorderInfo[channel.userName] = syncInfo{
		LatestMsgID:  latestID,
		LastSyncTime: createdAt,
	}

	err = db.From(channel.channelName).From(channel.userName).Save(&message)
	if err != nil {
		log.Println(err)
		return err
	}

	err = db.From(channel.channelName).Update(&info)
	if err != nil {
		return err
	}

	db.Commit()

	jsonChat, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	err = channel.pubsub.Publish(channel.channelName, jsonChat)
	if err != nil {
		return err
	}
	return nil
}

// Leave will leave the current channel(can not receive message anymore)
func (channel *ChatChannel) Leave() {
	channel.sub.Cancel()
	channel.db.Commit()
	channel.db.Close()
}

// On will assign handlers to NewMessage, NewPeer, PeerJoin, PeerLeave
func (channel *ChatChannel) On(event string, handler func(req []byte)) error {
	return nil
}

// GetPeers get current peers in the channel
func (channel *ChatChannel) GetPeers() []string {
	keys := make([]string, 0, len(channel.memberToPID))
	for key := range channel.memberToPID {
		keys = append(keys, key)
	}
	return keys
}

func (channel *ChatChannel) addMember(userName string) error {
	return nil
}

// subscribe a channel as topic using pubsub
func join(p *pubsub.PubSub, channelName string) (*pubsub.Subscription, error) {
	sub, err := p.Subscribe(channelName)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// get connection with all the peers in the network
func connect(host *host.Host, memberToAddrs map[string][]string) {
	for _, addrs := range memberToAddrs {
		for _, addr := range addrs {
			network.connectByMultiAddrString(host, addr)
		}
	}
}
