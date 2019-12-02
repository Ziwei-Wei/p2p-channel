package msglist

// PeerData is a state-based CRDT, all state comes from the Peer which publish it in network
type PeerData struct {
	PeerID       string `storm:"id"`
	PeerName     string `storm:"index"`
	P2PAddrs     []string
	LatestMsgID  int
	LastSyncTime int64
	MsgDict      map[int]bool
}

// Message for various messages should marchall []byte/structP{} to string in content field
type Message struct {
	ID        int `storm:"id"`
	SenderID  string
	AuthorID  string
	Content   string
	CreatedAt int64 `storm:"index"`
}
