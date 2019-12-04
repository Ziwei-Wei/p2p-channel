package channel

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/cyber-rhizome/msglist"
	util "github.com/cyber-rhizome/utility"
	"github.com/libp2p/go-libp2p-core/network"
)

// sync channel every 5 minute
func (c *ChatChannel) startSync(ctx context.Context) {
	log.Println("---> sync started")
	for {
		select {
		case <-ctx.Done():
			log.Printf("stop sync")
			return
		default:
			go c.sendAllPeerData()
			c.requestSyncFromPeers(ctx, 100)
			time.Sleep(20 * time.Second)
		}
	}
}

func (c *ChatChannel) sendAllPeerData() error {
	for _, list := range c.peerIDToMsgList {
		c.sendPeerData(list)
	}
	return nil
}

func (c *ChatChannel) sendPeerData(list *msglist.PeerMessageList) error {
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
	return c.sendPeerMessage(msg)
}

// sync channel msgs with a buffer limit
func (c *ChatChannel) requestSyncFromPeers(ctx context.Context, limit int) {
	msgCount := 0
	for _, list := range c.peerIDToMsgList {
		go c.requestSyncFromPeer(ctx, list, &msgCount)
		if msgCount > limit {
			for msgCount > limit/2 {
				if msgCount <= limit/2 {
					break
				}
			}
		}
	}
}

func (c *ChatChannel) requestSyncFromPeer(ctx context.Context, list *msglist.PeerMessageList, counter *int) error {
	if list.GetPeerID() != c.userID {
		missedMsgIDs, err := list.GetMissedMsgIDs()

		if err != nil {
			return err
		}

		if len(missedMsgIDs) > 0 {
			log.Println("syncMessages")
			*counter += len(missedMsgIDs)
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
			var peerIDToSync []string
			peerIDToSync = append(peerIDToSync, list.GetPeerID())
			for i := 0; i < len(latencies); i++ {
				if count < max && list.GetPeerID() != latencyToPeer[latencies[i]] {
					peerIDToSync = append(peerIDToSync, latencyToPeer[latencies[i]])
					count++
				}
			}

			for _, peerIDString := range peerIDToSync {
				err := c.connectPeer(ctx, util.StringsToMultiAddrs(list.P2PAddrs))
				if err != nil {
					log.Printf("Error: %v at requestSyncFromPeer, connectPeer", err)
					continue
				}

				peerID, err := util.StringToPeerID(peerIDString)
				if err != nil {
					log.Printf("Error: %v at requestSyncFromPeer, StringToPeerID(peerIDString)", err)
					continue
				}

				stream, err := c.host.NewStream(context.Background(), peerID, "/chat/1.0.0")
				if err != nil {
					log.Printf("Error: %v at requestSyncFromPeer, NewStream(context.Background(), peerID)", err)
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
				_, err = rw.Write(data)
				if err != nil {
					log.Printf("error: %v in requestSyncFromPeer, Write failed", err)
					continue
				}
				rw.Flush()
				for {
					data, err := rw.ReadBytes('\n')
					if err != nil {
						log.Printf("error: %v in requestSyncFromPeer, ReadBytes failed", err)
						continue
					}
					if data[0] == '\n' {
						break
					}
					var syncRes syncResponse
					err = json.Unmarshal(data[:len(data)-1], &syncRes)
					if err != nil {
						log.Printf("error: %v in requestSyncFromPeer, Unmarshal failed", err)
						continue
					}
					list.Add(syncRes.MissingMsg)
				}
				rw.Write([]byte{'\n'})
				rw.Flush()
				stream.Close()
				break
			}
			c.onSync(c, missedMsgIDs)
			*counter -= len(missedMsgIDs)
		}
	}
	return nil
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
			log.Printf("error: %v in handleSyncStream, ReadBytes failed", err)
			continue
		}

		var req syncRequest
		err = json.Unmarshal(data[:len(data)-1], &req)
		if err != nil {
			log.Printf("error: %v in handleSyncStream, Unmarshal failed", err)
			continue
		}

		if req.ChannelName != c.channelName {
			log.Printf("should be in the same channel")
			return
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
				log.Printf("error: %v in handleSyncStream, Marshal failed", err)
				continue
			}
			data = append(data, '\n')
			rw.Write(data)
			rw.Flush()
		}
		rw.Write([]byte{'\n'})
		rw.Flush()

		TTL = time.Now().Add(time.Minute)
		for {
			data, err := rw.ReadBytes('\n')
			if err != nil {
				log.Printf("error: %v in requestSyncFromPeer, ReadBytes failed", err)
				return
			}
			if data[0] == '\n' || time.Now().After(TTL) {
				stream.Close()
				return
			}
		}
	}
}
