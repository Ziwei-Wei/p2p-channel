package msglist

import (
	"errors"
	"log"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

/**
 * PeerMessageList is a CRDT, it is connected to a database.
 * It will send state to sync with other peers,
 * it will also send Add() to other peers to expand the message list.
 * It is a state based CRDT and will handle the operation based CRDT message list in db.
**/

// PeerMessageList for a single peer
type PeerMessageList struct {
	peerID       string
	PeerName     string
	P2PAddrs     []string
	LatestMsgID  int
	LastSyncTime int64
	db           storm.Node
}

// New create a new PeerMessageList, notice only peer id matters
func New(
	peerID string,
	peerName string,
	p2pAddrs []string,
	db storm.Node,
) (*PeerMessageList, error) {
	// check if peer exist
	var data PeerData
	err := db.One("PeerID", peerID, &data)
	if err != nil && err.Error() != "not found" {
		log.Printf("error: %v in Create, Find failed", err)
		return nil, err
	}
	// if not exist then create the peer, if exist load peer
	if data.PeerID == "" {
		data = PeerData{
			PeerID:       peerID,
			PeerName:     peerName,
			P2PAddrs:     p2pAddrs,
			LatestMsgID:  0,
			LastSyncTime: 0,
			MsgDict:      make(map[int]bool),
		}
		err = db.Save(&data)
		if err != nil {
			log.Printf("error: %v in Create, Save failed", err)
			return nil, err
		}

		msgList := &PeerMessageList{
			peerID:       peerID,
			PeerName:     peerName,
			P2PAddrs:     p2pAddrs,
			LatestMsgID:  0,
			LastSyncTime: 0,
			db:           db,
		}
		return msgList, nil
	}
	msgList := &PeerMessageList{
		peerID:       peerID,
		PeerName:     data.PeerName,
		P2PAddrs:     data.P2PAddrs,
		LatestMsgID:  data.LatestMsgID,
		LastSyncTime: data.LastSyncTime,
		db:           db,
	}
	return msgList, nil
}

// Add will add a message to the Peer's message list
func (p *PeerMessageList) Add(message Message) error {
	if message.AuthorID != p.peerID {
		return errors.New("wrong peer")
	}

	if message.ID > p.LatestMsgID {
		err := p.saveNewMessageToDB(&message)
		if err != nil {
			log.Printf("error: %v in Add, saveNewMessageToDB failed", err)
			return err
		}
		p.LatestMsgID = message.ID
	} else {
		err := p.saveOldMessageToDB(&message)
		if err != nil {
			log.Printf("error: %v in Add, saveOldMessageToDB failed", err)
			return err
		}
	}
	return nil
}

// FindMessage find by id in existing messages
func (p *PeerMessageList) FindMessage(messageIDs []int) ([]Message, error) {
	var messages []Message
	matchers := make([]q.Matcher, len(messageIDs))
	for i, messageID := range messageIDs {
		if messageID <= p.LatestMsgID {
			matchers[i] = q.Eq("ID", messageID)
		}
	}
	err := p.db.Select(q.Or(matchers...)).Find(&messages)
	return messages, err
}

// GetMissedMsgIDs for synchronization purpose
func (p *PeerMessageList) GetMissedMsgIDs() ([]int, error) {
	var data PeerData
	var missedMsgIDs []int
	err := p.db.One("PeerID", p.peerID, &data)
	if err != nil {
		log.Printf("error: %v in GetMissedMsgIDs, db.One failed", err)
		return nil, err
	}
	for ID, exist := range data.MsgDict {
		if exist == false {
			missedMsgIDs = append(missedMsgIDs, ID)
		}
	}
	return missedMsgIDs, nil
}

// GetPeerID retrieve the PeerID
func (p *PeerMessageList) GetPeerID() string {
	return p.peerID
}

// GetPeerData retrieve the PeerDataI
func (p *PeerMessageList) GetPeerData() PeerData {
	var data PeerData
	err := p.db.One("PeerID", p.peerID, &data)
	if err != nil {
		log.Printf("error: %v in GetPeerData, db.One failed", err)
	}
	return data
}

// Save msg list info to disk
func (p *PeerMessageList) Save() error {
	err := p.db.Update(&PeerData{
		PeerID:       p.peerID,
		PeerName:     p.PeerName,
		P2PAddrs:     p.P2PAddrs,
		LatestMsgID:  p.LatestMsgID,
		LastSyncTime: p.LastSyncTime,
	})
	if err != nil {
		log.Printf("error: %v in Save, Update failed", err)
		return err
	}
	return nil
}
