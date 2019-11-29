package channel

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

func (channel *chatChannel) syncChannel(limit int) {
	var allSync []userMessagesIndex
	channel.db.From(channel.channelName).All(&allSync)

	countChan := make(chan int)
	msgCount := 0
	for _, sync := range allSync {
		channel.syncUser(&sync, countChan)
		msgCount += <-countChan
		if msgCount > limit {
			for msgCount > limit/2 {
				msgCount += <-countChan
			}
		}
	}

}

func (channel *chatChannel) syncUser(userInfo *userMessagesIndex, countChan chan int) {
	if userInfo.userName != channel.userName {
		var missingMsgs []uint
		for index, missing := range userInfo.msgDict {
			if missing == true {
				missingMsgs = append(missingMsgs, index)
			}
		}

		if len(missingMsgs) > 0 {
			countChan <- len(missingMsgs)
			var peerMsgInfo peerMessageInfo
			channel.db.Find("userName", userInfo.userName, &peerMsgInfo)
			var syncPeers []string
			for _, info := range peerMsgInfo.syncInfo {
				if info.recorder == userInfo.userName ||
					info.latestMsgID >= userInfo.latestMsgID &&
						info.msgCount >= uint(len(userInfo.msgDict)) &&
						time.Now().Unix()-info.lastSyncTime <= 10*unixMinute {
					syncPeers = append(syncPeers, info.recorder)
				}
			}

			for _, peerName := range syncPeers {
				err := connectByMultiAddrStrings(channel.host, channel.memberToAddrs[peerName])
				if err != nil {
					continue
				}

				stream, err := (*channel.host).NewStream(context.Background(), peer.ID(channel.memberToPID[peerName]))
				if err != nil {
					continue
				}
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
				syncReq := syncRequest{
					userName:          userInfo.userName,
					missingMsgIndexes: missingMsgs,
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
					for _, chat := range syncRes.missingMsgs {
						channel.addChat(chat)
					}
					break
				}
				rw.Write([]byte{'\n'})
				rw.Flush()
				stream.Close()
				break
			}
			countChan <- -len(missingMsgs)
		}
	}
}

func (channel *chatChannel) sendSyncInfo(syncType string) {
	switch syncType {
	case "Messages":
		var allSync []userMessagesIndex
		channel.db.From(channel.channelName).All(&allSync)

		var latestIndexes map[string]uint
		for _, sync := range allSync {
			latestIndexes[sync.userName] = sync.latestMsgID
		}

		syncMsgs := syncMessages{
			sender:        channel.userName,
			latestIndexes: latestIndexes,
		}

		data, err := json.Marshal(&syncMsgs)
		if err != nil {
			println(err)
		}

		message := message{
			msgType: "syncMessages",
			data:    data,
		}

		channel.sendMessage(message)
		break

	case "Members":
		var allSync channelInfo
		channel.db.One("channelName", channel.channelName, &allSync)

		syncMsgs := syncMembers{
			sender:        channel.userName,
			memberToPID:   allSync.memberToPID,
			memberToAddrs: allSync.memberToAddrs,
		}

		data, err := json.Marshal(&syncMsgs)
		if err != nil {
			println(err)
		}

		message := message{
			msgType: "syncMembers",
			data:    data,
		}

		channel.sendMessage(message)
		break
	default:
		err := errors.New("sync type is not supported")
		println(err)
	}

}
