package channel

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	network "github.com/Ziwei-Wei/cyber-rhizome-host/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

// sync channel every 5 minute
func (channel *ChatChannel) sync() {
	for {
		channel.syncChannel(1000)
		time.Sleep(5 * time.Minute)
	}
}

// sync channel msgs with a buffer limit
func (channel *ChatChannel) syncChannel(limit int) {
	var allInfo []peerMessageInfo
	channel.db.From(channel.channelName).All(&allInfo)

	counter := make(chan int)
	msgCount := 0
	for _, sync := range allInfo {
		go channel.syncPeer(&sync, counter)
		msgCount += <-counter
		if msgCount > limit {
			for msgCount > limit/2 {
				msgCount += <-counter
			}
		}
	}

}

func (channel *ChatChannel) syncPeer(peerInfo *peerMessageInfo, counter chan int) {
	if peerInfo.PeerName != channel.userName {
		var missingMsgIDs []int
		for index, missing := range peerInfo.MsgDict {
			if missing == true {
				missingMsgIDs = append(missingMsgIDs, index)
			}
		}

		if len(missingMsgIDs) > 0 {
			counter <- len(missingMsgIDs)
			var peersToSync []string
			for peer, recorderInfo := range peerInfo.RecordersInfo {
				if recorderInfo.LatestMsgID >= peerInfo.RecordersInfo[channel.userName].LatestMsgID &&
					time.Now().Unix()-recorderInfo.LastSyncTime <= 10*unixMinute {
					peersToSync = append(peersToSync, peer)
				}
			}

			for _, peerName := range peersToSync {
				err := network.ConnectByMultiAddrStrings(channel.host, channel.memberToAddrs[peerName])
				if err != nil {
					continue
				}

				stream, err := (*channel.host).NewStream(context.Background(), peer.ID(channel.memberToPID[peerName]))
				if err != nil {
					continue
				}
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
				syncReq := syncRequest{
					UserName:          peerInfo.PeerName,
					MissingMsgIndexes: missingMsgIDs,
				}
				data, err := json.Marshal(&syncReq)
				if err != nil {
					continue
				}
				data = append(data, '\n')
				rw.Write(data)
				rw.Flush()
				for {
					data, err := rw.ReadBytes('\n')
					if err != nil {
						print(err)
					}
					if data[0] == '\n' {
						break
					}
					var syncRes syncResponse
					err = json.Unmarshal(data[:len(data)-1], &syncRes)
					if err != nil {
						print(err)
					}
					for _, msg := range syncRes.MissingMsgs {
						channel.saveMessage(msg)
					}
					break
				}
				rw.Write([]byte{'\n'})
				rw.Flush()
				stream.Close()
				break
			}
			counter <- -len(missingMsgIDs)
		}
	}
}

func (channel *ChatChannel) sendSyncInfo(syncType string) {
	switch syncType {
	case "Messages":
		var allSync []peerMessageInfo
		err := channel.db.From(channel.channelName).All(&allSync)
		if err != nil {
			log.Println(err)
			break
		}

		var LatestIDs map[string]int
		for _, sync := range allSync {
			LatestIDs[sync.PeerName] = sync.RecordersInfo[channel.userName].LatestMsgID
		}

		syncMsgs := syncMessages{
			Sender:    channel.userName,
			LatestIDs: LatestIDs,
		}

		data, err := json.Marshal(&syncMsgs)
		if err != nil {
			log.Println(err)
			break
		}

		message := peerMessage{
			MsgType: "syncMessages",
			Data:    data,
		}

		channel.sendPeerMessage(message)
		break

	case "Members":
		var allSync channelInfo
		err := channel.db.One("channelName", channel.channelName, &allSync)
		if err != nil {
			log.Println(err)
			break
		}

		syncMsgs := syncMembers{
			Sender:        channel.userName,
			MemberToPID:   allSync.MemberToPID,
			MemberToAddrs: allSync.MemberToAddrs,
		}

		data, err := json.Marshal(&syncMsgs)
		if err != nil {
			log.Println(err)
			break
		}

		message := peerMessage{
			MsgType: "syncMembers",
			Data:    data,
		}

		channel.sendPeerMessage(message)
		break
	default:
		err := errors.New("sync type is not supported")
		println(err)
	}

}

func (channel *ChatChannel) sendPeerMessage(message peerMessage) error {
	data, err := json.Marshal(&message)
	if err != nil {
		log.Println(err)
		return err
	}
	err = channel.pubsub.Publish(channel.channelName, data)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
