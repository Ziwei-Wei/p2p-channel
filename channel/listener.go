package channel

import (
	"context"
	"encoding/json"
)

func (channel *ChatChannel) listenToPeers() error {
	for {
		// listen
		msg, _ := channel.sub.Next(context.Background())

		// unmarshall
		var message peerMessage
		json.Unmarshal(msg.GetData(), &message)

		switch message.MsgType {
		case "chatMessage":
			go channel.handleChatMessage(message.Data)
			break
		case "syncMessages":
			go channel.handleSyncMessages(message.Data)
			break
		case "syncMembers":
			go channel.handleSyncMembers(message.Data)
			break
		default:
			println("message type is not supported")
		}
	}
	return nil
}
