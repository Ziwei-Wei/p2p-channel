package space

/*
Space is a group of pubsub channels sharing the same private p2p network
with synchronization abilities and eventual message consistency.
It can be used to support various applications such as chatroom, co-edit doc, p2p database, etc.
*/

// Space is a persistent pubsub channel(have to link to a local db)
type Space interface {
	// will send message to all members of the channel
	Send(channel, data []byte) error

	// will leave the current channel(can not receive message anymore)
	Logout()

	// will assign handlers to NewMessage, NewPeer, PeerJoin, PeerLeave
	On(event string, handler func(req []byte)) error

	// get current peers name list in the space
	GetPeers() []string
}
