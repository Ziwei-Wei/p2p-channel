package channel

import (
	"log"
	"time"

	network "github.com/Ziwei-Wei/cyber-rhizome-host/network"
	host "github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	unixWeek   int64 = 604800
	unixDay    int64 = 86400
	unixHour   int64 = 3600
	unixMinute int64 = 60
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
func connect(host *host.Host, memberToAddrs map[string][]string) {
	for _, addrs := range memberToAddrs {
		for _, addr := range addrs {
			network.ConnectByMultiAddrString(host, addr)
		}
	}
}

func (channel *ChatChannel) saveMessage(message message) error {
	db, err := channel.db.Begin(true)
	if err != nil {
		return err
	}
	defer db.Rollback()

	var info peerMessageInfo
	err = db.From(channel.channelName).One("PeerName", message.Author, &info)
	if err != nil {
		log.Println(err)
		return err
	}
	if info.MsgDict[message.ID] == false {
		err = db.From(channel.channelName).From(channel.userName).Save(&message)
		if err != nil {
			log.Println(err)
			return err
		}

		info.MsgDict[message.ID] = true
		if message.ID > info.RecordersInfo[channel.userName].LatestMsgID {
			info.RecordersInfo[channel.userName] = syncInfo{
				LatestMsgID:  message.ID,
				LastSyncTime: time.Now().Unix(),
			}
		}

		err = db.From(channel.channelName).Update(&info)
		if err != nil {
			return err
		}
	}
	return db.Commit()
}

func (channel *ChatChannel) getPeerLatestMsgID(peerName string) (int, error) {
	var info peerMessageInfo
	err := channel.db.
		From(channel.channelName).
		One("PeerName", peerName, &info)
	if err != nil {
		return -1, err
	}
	return info.RecordersInfo[peerName].LatestMsgID, nil
}

func (channel *ChatChannel) updatePeerLatestMsgID(peerName string, newMsgID int, lastSyncTime int64) error {
	err := channel.db.
		From(channel.channelName).
		UpdateField(&peerMessageInfo{PeerName: peerName}, "RecorderInfo", newMsgID)

	if err != nil {
		return err
	}
	return nil
}

func (channel *ChatChannel) getAllLatestMsgIDs() ([]peerMessageInfo, error) {
	var info []peerMessageInfo
	err := channel.db.
		From(channel.channelName).
		All(&info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
