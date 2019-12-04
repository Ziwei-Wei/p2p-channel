package channel

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// should be dynamic in future
func (c *ChatChannel) sendHeartBeat(ctx context.Context) {
	time.Sleep(time.Second)
	c.sendStateInfo()
	c.state = peerCONNECTED
	for {
		select {
		case <-ctx.Done():
			log.Printf("stop heart beat")
			return
		default:
			c.sendStateInfo()
			time.Sleep(5 * time.Second)
		}
	}
}

func (c *ChatChannel) sendStateInfo() error {
	state := syncRawState{
		P2pAddrs: c.peerIDToMsgList[c.userID].P2PAddrs,
		State:    c.state,
	}
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("error: %v at sendStateInfo, Marshal(state)", err)
		return err
	}
	err = c.sendPeerMessage(pubsubRawMessage{
		MsgType: stateMESSAGE,
		Data:    data,
	})
	if err != nil {
		log.Printf("error: %v at sendStateInfo, sendPeerMessage", err)
		return err
	}
	return nil
}

func (c *ChatChannel) handleStates(ctx context.Context) {
	for {
		select {
		case state := <-c.stateQueue:
			switch state.State {
			case peerJOIN:
				c.onPeerJoin(c, state.SenderID)
				break
			case peerCONNECTED:
				if c.peerIDToMsgList[state.SenderID] == nil {
					c.AddPeer(state.SenderID, "unknown", state.P2pAddrs)
					c.onNewPeer(c, state.SenderID)
				}
				c.onPeerConnected(c, state.SenderID)
				break
			case peerLEAVE:
				c.onPeerLeave(c, state.SenderID)
				break
			default:
				log.Printf("not supported state: %s", state.State)
			}
		case <-ctx.Done():
			log.Printf("exit handleStates")
			return
		}
	}
}
