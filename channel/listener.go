package channel

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

func (channel *chatChannel) listenToGUI(w http.ResponseWriter, r *http.Request) {
	if channel.ready != true {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		channel.gui = conn
	}
}

func (channel *chatChannel) listenToPeers() error {
	for {
		// listen
		msg, _ := channel.sub.Next(context.Background())

		// unmarshall
		var message message
		json.Unmarshal(msg.GetData(), &message)

		switch message.msgType {
		case "chatMessage":
			go channel.handleChatMessage(message.data)
			break
		case "syncMessages":
			go channel.handleSyncMessages(message.data)
			break
		case "syncMembers":
			go channel.handleSyncMembers(message.data)
			break
		default:
			println("message type is not supported")
		}
	}
	return nil
}
