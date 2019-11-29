package spce

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
}

func (core *RhizomeCore) listenToGUI(w http.ResponseWriter, r *http.Request) {
	if core.ready != true {
		var upgrader = websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		go core.reader(conn)
	}
}

func (core *RhizomeCore) reader(conn *websocket.Conn) error {
	defer conn.Close()
	for {
		var req message
		err := conn.ReadJSON(&req)
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", req)

		switch req.msgType {
		case "login":
			break
		case "logout":
			break
		case "sendMessage":
			var chat sendChatMessage
			json.Unmarshal(req.data, &chat)
			core.channels[chat.channel].sendChat(chat.message)
			break
		case "syncMessages":
			var syncReq syncRequest
			json.Unmarshal(req.data, &syncReq)

			syncRes := syncResponse{
				channelName: syncReq.channelName,
				userName:    syncReq.userName,
				missingMsgs: make([]chatMessage, 0, len(syncReq.missingMsgIndexes)),
			}
			for _, id := range syncReq.missingMsgIndexes {
				var chat chatMessage
				err := core.db.From(syncReq.channelName).From(syncReq.userName).Find("id", id, &chat)
				if err != nil {
					log.Println(err)
					continue
				}
				syncRes.missingMsgs = append(syncRes.missingMsgs, chat)
			}
			data, err := json.Marshal(&syncRes)
			if err != nil {
				log.Println(err)
				break
			}
			res := message{
				msgType: "syncResponse",
				data:    data,
			}
			conn.WriteJSON(res)
			break
		case "createDoc":
			break
		case "sendEdit":
			break
		case "syncEdit":
			break
		default:
			log.Println("message type is not supported")
		}
	}
	return nil
}
