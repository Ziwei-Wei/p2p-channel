package channel

import (
	"context"
	"encoding/json"
	"log"
)

func (c *ChatChannel) startListening(ctx context.Context) {
	go c.listenToPubSub(ctx)
	go c.listenToPubSub(ctx)
}

func (c *ChatChannel) listenToPubSub(ctx context.Context) error {
	for {
		// listen
		next, err := c.sub.Next(ctx)
		if err != nil {
			log.Printf("error: %v at listenToMessages, Next(ctx)", err)
			return err
		}

		// unmarshall
		var msg pubsubRawMessage
		json.Unmarshal(next.GetData(), &msg)
		if err != nil {
			log.Printf("error: %v at listenToMessages, Unmarshal(next.GetData(), &msg)", err)
			continue
		}

		select {
		case c.pubsubQueue <- pubsubMessage{
			MsgType: msg.MsgType,
			Sender:  next.GetFrom().Pretty(),
			Data:    msg.Data,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (channel *ChatChannel) listenToPeers(ctx context.Context) {
	// listening to messages
	log.Println("---> listener started")
	for {
		select {
		case msg := <-channel.pubsubQueue:
			switch msg.MsgType {
			case chatMESSAGE:
				var data chatMessage
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					log.Printf("error: %v at listenToPeers(), message", err)
					break
				}
				log.Printf("incoming chatMessage: %v", data)
				channel.chatQueue <- data
				break
			case syncMESSAGE:
				var data syncMessage
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					log.Printf("error: %v at listenToPeers(), syncMessages", err)
					break
				}
				log.Printf("incoming syncMessages: %v", data)
				channel.syncMsgQueue <- data
				break
			case stateMESSAGE:
				var data syncState
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					log.Printf("error: %v at listenToPeers(), state", err)
					break
				}
				log.Printf("incoming State: %v", data)
				channel.stateQueue <- data
				break
			default:
				println("message type is not supported")

			}
		case <-ctx.Done():
			return
		}
	}
}
