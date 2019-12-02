package channel

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/Ziwei-Wei/cyber-rhizome-host/msglist"
	util "github.com/Ziwei-Wei/cyber-rhizome-host/utility"
)

const unixMinute int64 = 60

// sync channel every 5 minute
func (c *ChatChannel) syncChannel(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
		default:
			go c.sendSyncInfo()
			go c.sendSyncMsgInfoToPeers(100)
			time.Sleep(5 * time.Minute)
		}
	}
}

// sync channel msgs with a buffer limit
func (c *ChatChannel) sendSyncMsgInfoToPeers(limit int) {
	counter := make(chan int)
	msgCount := 0
	for _, list := range c.peerIDToMsgList {
		go c.sendSyncMsgInfoToPeer(list, counter)
		msgCount += <-counter
		if msgCount > limit {
			for msgCount > limit/2 {
				msgCount += <-counter
			}
		}
	}

}

func (c *ChatChannel) sendSyncMsgInfoToPeer(list *msglist.PeerMessageList, counter chan int) error {
	if list.GetPeerID() != c.userID {
		missedMsgIDs, err := list.GetMissedMsgIDs()

		if err != nil {
			return err
		}

		if len(missedMsgIDs) > 0 {
			counter <- len(missedMsgIDs)
			latencyToPeer := make(map[int]string)
			latencies := make([]int, len(c.peerIDToMsgList))
			// find 2 peer with least latency + author
			for id := range c.peerIDToMsgList {
				peerID, _ := util.StringToPeerID(id)
				t := c.host.Peerstore().LatencyEWMA(peerID).Milliseconds()
				latencies = append(latencies, int(t))
				latencyToPeer[int(t)] = id
			}

			count := 1
			max := 3
			sort.Ints(latencies)
			peerIDToSync := make([]string, max)
			peerIDToSync[0] = list.GetPeerID()
			for i := 0; i < len(latencies); i++ {
				if count < max {
					peerIDToSync[i+1] = latencyToPeer[latencies[i]]
					count++
				}
			}

			for _, peerIDString := range peerIDToSync {
				err := c.connectPeer(util.StringsToMultiAddrs(list.P2PAddrs))
				if err != nil {
					log.Printf("Error: %v at syncWithPeer, connectPeer", err)
					continue
				}

				peerID, err := util.StringToPeerID(peerIDString)
				if err != nil {
					log.Printf("Error: %v at syncWithPeer, StringToPeerID(peerIDString)", err)
					continue
				}

				stream, err := c.host.NewStream(context.Background(), peerID)
				if err != nil {
					log.Printf("Error: %v at syncWithPeer, NewStream(context.Background(), peerID)", err)
					continue
				}

				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
				syncReq := syncRequest{
					ChannelName:   c.channelName,
					TargerID:      list.GetPeerID(),
					MissingMsgIDs: missedMsgIDs,
				}
				data, err := json.Marshal(&syncReq)
				if err != nil {
					continue
				}
				data = append(data, '\n')
				rw.Write(data)
				rw.Flush()
				for {
					data, err := rw.ReadBytes('\n')
					if err != nil {
						print(err)
					}
					if data[0] == '\n' {
						break
					}
					var syncRes syncResponse
					err = json.Unmarshal(data[:len(data)-1], &syncRes)
					if err != nil {
						print(err)
					}
					list.Add(syncRes.MissingMsg)
					break
				}
				rw.Write([]byte{'\n'})
				rw.Flush()
				stream.Close()
				break
			}
			counter <- -len(missedMsgIDs)
		}
	}
	return nil
}

// should be dynamic in future
func (c *ChatChannel) sendHeartBeat() {
	for {
		c.sendStateInfo()
		time.Sleep(5 * time.Second)
	}
}

func (c *ChatChannel) sendStateInfo() error {
	state := syncRawState{
		State: c.state,
	}
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("error: %v at sendStateInfo, Marshal(state)", err)
		return err
	}
	err = c.sendPeerMessage(pubsubRawMessage{
		MsgType: "State",
		Data:    data,
	})
	if err != nil {
		log.Printf("error: %v at sendStateInfo, sendPeerMessage", err)
		return err
	}
	return nil
}

func (c *ChatChannel) sendSyncInfo() error {
	list := c.peerIDToMsgList[c.userID]
	syncMsg := syncRawMessage{
		TargerID:     list.GetPeerID(),
		TargetName:   list.PeerName,
		P2pAddrs:     list.P2PAddrs,
		LastSyncTime: list.LastSyncTime,
		LatestMsgID:  list.LatestMsgID,
	}
	data, err := json.Marshal(&syncMsg)
	if err != nil {
		log.Printf("error: %v at sendSyncInfo, Marshal(&message)", err)
		return err
	}
	msg := pubsubRawMessage{
		MsgType: syncMESSAGE,
		Data:    data,
	}
	c.sendPeerMessage(msg)
	return nil
}

func (c *ChatChannel) sendPeerMessage(message pubsubRawMessage) error {
	data, err := json.Marshal(&message)
	if err != nil {
		log.Printf("error: %v at sendPeerMessage, Marshal(&message)", err)
		return err
	}

	err = c.pubsub.Publish(c.channelName, data)
	if err != nil {
		log.Printf("error: %v at sendPeerMessage, Publish(c.cName, data)", err)
		return err
	}
	return nil
}
