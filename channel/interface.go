package channel

/*
Channel is a pubsub channel
with synchronization abilities and eventual message consistency.
It can be used to support various applications such as chatroom, co-edit doc, p2p database, etc.
*/

// Channel is a persistent pubsub channel(have to link to a local db)
type Channel interface {
	// will send message to all members of the channel
	Send(data []byte) error

	// will leave the current channel(can not receive message anymore)
	Leave()

	// will assign handlers to NewMessage, NewPeer, PeerJoin, PeerLeave
	On(event string, handler func(req []byte)) error

	// get current peers name list in the channel
	GetPeers() []string
}
