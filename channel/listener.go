package channel

import (
	"context"
	"encoding/json"
	"log"
)

func (c *ChatChannel) startListeners(ctx context.Context) {
	log.Println("---> listener started")
	go c.listenToPubSub(ctx)
	go c.listenToPeers(ctx)
}

func (c *ChatChannel) listenToPubSub(ctx context.Context) {
	for {
		// listen
		next, err := c.sub.Next(ctx)
		if err != nil {
			if err.Error() == "subscription cancelled by calling sub.Cancel()" || err.Error() == "context canceled" {
				log.Printf("exit listenToPubSub")
				return
			}
			log.Printf("error: %v at listenToPubSub, Next(ctx)", err)
			continue
		}

		if next.GetFrom().Pretty() == c.userID {
			continue
		}

		// unmarshall
		var msg pubsubRawMessage
		json.Unmarshal(next.GetData(), &msg)
		if err != nil {
			log.Printf("error: %v at listenToPubSub, Unmarshal(next.GetData(), &msg)", err)
			continue
		}

		// if c.peerIDToMsgList[c.userID].PeerName == "tester0" {
		// 	log.Println(next.GetFrom().Pretty())
		// 	log.Println(msg.MsgType)
		// }

		select {
		case c.pubsubQueue <- pubsubMessage{
			MsgType: msg.MsgType,
			Sender:  next.GetFrom().Pretty(),
			Data:    msg.Data,
		}:
		case <-ctx.Done():
			log.Printf("exit listenToPubSub")
			return
		}
	}
}

func (c *ChatChannel) listenToPeers(ctx context.Context) {
	for {
		select {
		case msg := <-c.pubsubQueue:
			switch msg.MsgType {
			case chatMESSAGE:
				var data chatMessage
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					log.Printf("error: %v at listenToPeers(), message", err)
					break
				}
				data.AuthorID = msg.Sender
				c.chatQueue <- data
				break
			case syncMESSAGE:
				var data syncMessage
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					log.Printf("error: %v at listenToPeers(), syncMessages", err)
					break
				}
				data.SenderID = msg.Sender
				c.syncMsgQueue <- data
				break
			case stateMESSAGE:
				var data syncState
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					log.Printf("error: %v at listenToPeers(), state", err)
					break
				}
				data.SenderID = msg.Sender
				c.stateQueue <- data
				break
			default:
				log.Printf("message type %s is not supported", msg.MsgType)
			}
		case <-ctx.Done():
			log.Printf("exit listenToPeers")
			return
		}
	}
}
