package channel

import (
	"bufio"
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
)

func (channel *chatChannel) handleChatMessage(data json.RawMessage) error {
	// save chat to db
	var chat chatMessage
	json.Unmarshal(data, &chat)
	channel.db.From(channel.channelName).From(chat.author).Save(&chat)

	// sync to db
	var sync userMessagesIndex
	channel.db.From(channel.channelName).Find("userName", chat.author, &sync)
	if sync.latestMsgID < chat.id {
		sync.latestMsgID = chat.id
	}
	sync.msgDict[chat.id] = true
	channel.db.From(channel.channelName).Update(&sync)

	// send to gui
	channel.mutex.Lock()
	channel.gui.WriteJSON(&chat)
	channel.mutex.Unlock()
	return nil
}

func (channel *chatChannel) handleSyncMessages(data json.RawMessage) error {
	var syncMsgs syncMessages
	json.Unmarshal(data, &syncMsgs)

	var allSync []userMessagesIndex
	channel.db.From(channel.channelName).All(&allSync)

	for _, sync := range allSync {
		for i := sync.latestMsgID + 1; i <= syncMsgs.latestIndexes[sync.userName]; i++ {
			sync.msgDict[i] = false
		}
		channel.db.From(channel.channelName).Save(&sync)
	}

	return nil
}

func (channel *chatChannel) handleSyncMembers(data json.RawMessage) error {
	var syncInfo syncMembers
	json.Unmarshal(data, &syncInfo)

	var localInfo channelInfo
	channel.db.One("channelName", channel.channelName, &localInfo)

	if len(localInfo.memberToPID) < len(syncInfo.memberToPID) {
		for userName, peerID := range syncInfo.memberToPID {
			if localInfo.memberToPID[userName] == "" {
				localInfo.memberToPID[userName] = peerID
			}
		}
		for userName, addrs := range syncInfo.memberToAddrs {
			if localInfo.memberToAddrs[userName] == nil {
				localInfo.memberToAddrs[userName] = addrs
			}
		}
	}
	localInfo.memberToPID[syncInfo.sender] = syncInfo.memberToPID[syncInfo.sender]
	localInfo.memberToAddrs[syncInfo.sender] = syncInfo.memberToAddrs[syncInfo.sender]

	return nil
}

func (channel *chatChannel) handleSyncStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	var missingMsgs []chatMessage
	TTL := time.Now().Add(time.Minute)
	for {
		if time.Now().After(TTL) {
			stream.Close()
			return
		}
		data, err := rw.ReadBytes('\n')
		if err != nil {
			print(err)
			continue
		}
		var req syncRequest
		err = json.Unmarshal(data[:len(data)-1], &req)
		if err != nil {
			print(err)
			continue
		}
		for _, index := range req.missingMsgIndexes {
			var chat chatMessage
			err := channel.db.From(req.channelName).From(req.userName).Find("id", index, &chat)
			if err != nil {
				print(err)
				continue
			}
			missingMsgs = append(missingMsgs, chat)
		}
		var res syncResponse = syncResponse{
			channelName: req.channelName,
			userName:    req.userName,
			missingMsgs: missingMsgs,
		}
		resRaw, err := json.Marshal(&res)
		resRaw = append(resRaw, '\n')
		rw.Write(resRaw)
		rw.Flush()
		TTL = time.Now().Add(time.Minute)
		for {
			if data[0] == '\n' || time.Now().After(TTL) {
				stream.Close()
				return
			}
		}
	}
}
