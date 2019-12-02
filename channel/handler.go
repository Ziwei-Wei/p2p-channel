package channel

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Ziwei-Wei/cyber-rhizome-host/msglist"
	"github.com/libp2p/go-libp2p-core/network"
)

func (c *ChatChannel) handleStates(ctx context.Context) {
	log.Println("--> state queue handler started")
	for {
		select {
		case state := <-c.stateQueue:
			switch state.State {
			case "Join":
				log.Printf("peer %s joined", state.SenderID)
				c.onPeerJoin(state.SenderID)
				break
			case "Alive":
				log.Printf("peer %s Alive", state.SenderID)
				c.onPeerJoin(state.SenderID)
				break
			case "Leave":
				log.Printf("peer %s left", state.SenderID)
				c.onPeerLeave(state.SenderID)
				break
			default:
				log.Println("not supported state")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *ChatChannel) handleMessages(ctx context.Context) {
	log.Println("start to handle new messages")
	for {
		select {
		case msg := <-c.chatQueue:
			go c.saveChatMessage(msg)
			c.onNewMessage(msg)
		case <-ctx.Done():
			return
		}
	}
}

func (c *ChatChannel) handleSync(ctx context.Context) {
	log.Println("--> sync handler started")
	for {
		select {
		case syncMsg := <-c.syncMsgQueue:
			list := c.peerIDToMsgList[syncMsg.TargerID]
			if syncMsg.SenderID == list.GetPeerID() {
				list.PeerName = syncMsg.TargetName
				list.P2PAddrs = syncMsg.P2pAddrs
				list.LatestMsgID = syncMsg.LatestMsgID
				list.LastSyncTime = syncMsg.LastSyncTime
				list.Save()
			} else {
				if syncMsg.LatestMsgID > list.LatestMsgID {
					list.LatestMsgID = syncMsg.LatestMsgID
					list.Save()
				}
			}
		case <-ctx.Done():
			return
		}
	}
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

func (c *ChatChannel) handleSyncStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
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

		if req.ChannelName != c.channelName {
			continue
		}

		list := c.peerIDToMsgList[req.TargerID]
		msgs, err := list.FindMessage(req.MissingMsgIDs)
		for _, msg := range msgs {
			rawRes := syncResponse{
				ChannelName: c.channelName,
				TargerID:    req.TargerID,
				MissingMsg:  msg,
			}
			data, err := json.Marshal(&rawRes)
			if err != nil {

			}
			data = append(data, '\n')
			rw.Write(data)
			rw.Flush()
		}

		TTL = time.Now().Add(time.Minute)
		for {
			if data[0] == '\n' || time.Now().After(TTL) {
				stream.Close()
				return
			}
		}
	}
}
