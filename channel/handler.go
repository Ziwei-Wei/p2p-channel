package channel

import (
	"context"
	"log"
)

func (c *ChatChannel) startHandlers(ctx context.Context) {
	log.Println("---> handlers started")
	go c.handleMessages(ctx)
	go c.handleSync(ctx)
	go c.handleStates(ctx)
}

func (c *ChatChannel) handleMessages(ctx context.Context) {
	for {
		select {
		case msg := <-c.chatQueue:
			go c.saveChatMessage(msg)
			c.onNewMessage(c, msg)
		case <-ctx.Done():
			log.Printf("exit handleMessages")
			return
		}
	}
}

func (c *ChatChannel) handleSync(ctx context.Context) {
	for {
		select {
		case syncMsg := <-c.syncMsgQueue:
			list := c.peerIDToMsgList[syncMsg.TargerID]
			if list == nil {
				c.AddPeer(syncMsg.TargerID, syncMsg.TargetName, syncMsg.P2pAddrs)
				list = c.peerIDToMsgList[syncMsg.TargerID]
				c.onNewPeer(c, syncMsg.TargerID)
			}

			if syncMsg.LatestMsgID > list.LatestMsgID {
				data := list.GetPeerData()
				for i := data.LatestMsgID + 1; i < syncMsg.LatestMsgID+1; i++ {
					data.MsgDict[i] = false
				}
				data.PeerName = syncMsg.TargetName
				data.P2PAddrs = syncMsg.P2pAddrs
				data.LatestMsgID = syncMsg.LatestMsgID
				data.LastSyncTime = syncMsg.LastSyncTime
				list.UpdatePeerData(data)
			} else {
				if syncMsg.SenderID == list.GetPeerID() {
					list.PeerName = syncMsg.TargetName
					list.P2PAddrs = syncMsg.P2pAddrs
					list.LatestMsgID = syncMsg.LatestMsgID
					list.LastSyncTime = syncMsg.LastSyncTime
					list.Save()
				}
			}
		case <-ctx.Done():
			log.Printf("exit handleSync")
			return
		}
	}
}
