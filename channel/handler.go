package channel

import (
	"bufio"
	"encoding/json"
	"log"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
)

func (channel *ChatChannel) handleChatMessage(data json.RawMessage) error {
	db, err := channel.db.Begin(true)
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Rollback()

	// save chat to db
	var chat message
	json.Unmarshal(data, &chat)
	db.From(channel.channelName).From(chat.Author).Save(&chat)
	if err != nil {
		log.Println(err)
		return err
	}

	// save info to db
	var info peerMessageInfo
	db.From(channel.channelName).Find("PeerName", chat.Author, &info)
	if err != nil {
		log.Println(err)
		return err
	}

	var latestID int = info.RecordersInfo[chat.Author].LatestMsgID
	if latestID < chat.ID {
		latestID = chat.ID
	}
	info.RecordersInfo[chat.Author] = syncInfo{
		LatestMsgID:  latestID,
		LastSyncTime: time.Now().Unix(),
	}
	info.MsgDict[chat.ID] = true

	db.From(channel.channelName).Update(&info)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// !!!!!!!!!!!!!!! bug may be here
func (channel *ChatChannel) handleSyncMessages(data json.RawMessage) error {
	var syncMsg syncMessages
	json.Unmarshal(data, &syncMsg)

	var allInfo []peerMessageInfo
	channel.db.From(channel.channelName).All(&allInfo)

	for _, info := range allInfo {
		info.RecordersInfo[syncMsg.Sender] = syncInfo{
			LatestMsgID:  syncMsg.LatestIDs[info.PeerName],
			LastSyncTime: time.Now().Unix(),
		}
		channel.db.From(channel.channelName).Update(&info)
	}

	return nil
}

func (channel *ChatChannel) handleSyncMembers(data json.RawMessage) error {
	var syncMsg syncMembers
	json.Unmarshal(data, &syncMsg)

	var localInfo channelInfo
	err := channel.db.One("ChannelName", channel.channelName, &localInfo)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(localInfo.MemberToPID) < len(syncMsg.MemberToPID) {
		for UserName, peerID := range syncMsg.MemberToPID {
			if localInfo.MemberToPID[UserName] == "" {
				localInfo.MemberToPID[UserName] = peerID
			}
		}
		for UserName, addrs := range syncMsg.MemberToAddrs {
			if localInfo.MemberToAddrs[UserName] == nil {
				localInfo.MemberToAddrs[UserName] = addrs
			}
		}
	}
	localInfo.MemberToPID[syncMsg.Sender] = syncMsg.MemberToPID[syncMsg.Sender]
	localInfo.MemberToAddrs[syncMsg.Sender] = syncMsg.MemberToAddrs[syncMsg.Sender]

	err = channel.db.Update(localInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (channel *ChatChannel) handleSyncStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	var missingMsgs []message
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
		for _, index := range req.MissingMsgIndexes {
			var chat message
			err := channel.db.From(req.ChannelName).From(req.UserName).Find("id", index, &chat)
			if err != nil {
				print(err)
				continue
			}
			missingMsgs = append(missingMsgs, chat)
		}
		var res syncResponse = syncResponse{
			ChannelName: req.ChannelName,
			UserName:    req.UserName,
			MissingMsgs: missingMsgs,
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
