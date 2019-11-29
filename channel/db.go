package channel

import (
	"encoding/json"
)

/* db hierarchy */

/* channel level */
// one per channel, recore channel meta data
type channelInfo struct {
	ChannelName   string `storm:"id"`
	ChannelType   string `storm:"index"`
	MemberToPID   map[string]string
	MemberToAddrs map[string][]string
}

// one per peer, record local messages in the channel
type peerMessageInfo struct {
	PeerName     string `storm:"id"`
	MsgDict      map[int]bool
	RecorderInfo map[string]syncInfo
}

type syncInfo struct {
	LatestMsgID  int
	LastSyncTime int64
}

/* channel/user level */
// many chat to one user, recore one chat message
type message struct {
	ID        int `storm:"id"`
	Author    string
	Content   json.RawMessage
	CreatedAt int64 `storm:"index"`
}

type chat struct {
	Content string
}

// state: INPUTTING, INSERT, DELETE
// comes with sync, For rope structure
type doc struct {
	Action   string
	Position int
	Length   int
	Edit     string
}

/* communication between channels */
// data sent between users in the channel
type peerMessage struct {
	MsgType string
	Data    json.RawMessage
}

/* chat */

// ask for missing messages
type syncRequest struct {
	ChannelName       string
	UserName          string
	MissingMsgIndexes []int
}

// response for asking
type syncResponse struct {
	ChannelName string
	UserName    string
	MissingMsgs []message
}

// sync when add new member or logged in the channel
type syncMembers struct {
	Sender        string
	MemberToPID   map[string]string
	MemberToAddrs map[string][]string
}

// sync every 5 minute, or when user log in
type syncMessages struct {
	Sender        string
	LatestIndexes map[string]int
}

/* doc */
