package utility

import (
	"fmt"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// BuildP2PAddrsFromString ...
func BuildP2PAddrsFromString(peerID string, multiAddrs []string) []string {
	p2pAddrs := make([]string, len(multiAddrs))
	for i, multiAddr := range multiAddrs {
		p2pAddrs[i] = fmt.Sprintf("%s/p2p/%s", multiAddr, peerID)

	}
	return p2pAddrs
}

// BuildP2PAddrs ...
func BuildP2PAddrs(peerID peer.ID, multiAddrs []multiaddr.Multiaddr) []string {
	p2pAddrs := make([]string, len(multiAddrs))
	for i, multiAddr := range multiAddrs {
		p2pAddrs[i] = fmt.Sprintf("%s/p2p/%s", multiAddr.String(), peerID.Pretty())

	}
	return p2pAddrs
}

// SplitP2PAddrs ...
func SplitP2PAddrs(p2pAddrs []multiaddr.Multiaddr) map[string][]string {
	PeerIDToP2PAddrs := make(map[string][]string)
	for _, p2pAddr := range p2pAddrs {
		multiAddr, peerID := peer.SplitAddr(p2pAddr)
		PeerIDToP2PAddrs[peerID.Pretty()] = append(PeerIDToP2PAddrs[peerID.Pretty()], multiAddr.String())
	}
	return PeerIDToP2PAddrs
}

// MultiaddrsToStrings ...
func MultiaddrsToStrings(multiAddrs []multiaddr.Multiaddr) []string {
	strs := make([]string, len(multiAddrs))
	for i, multiAddr := range multiAddrs {
		strs[i] = multiAddr.String()
	}
	return strs
}

// MultiaddrToString ...
func MultiaddrToString(multiAddr multiaddr.Multiaddr) string {
	return multiAddr.String()
}

// StringsToMultiAddrs ...
func StringsToMultiAddrs(strs []string) []multiaddr.Multiaddr {
	multiAddrs := make([]multiaddr.Multiaddr, len(strs))
	for i, str := range strs {
		multiAddrs[i] = StringToMultiAddr(str)
	}
	return multiAddrs
}

// StringToMultiAddr ...
func StringToMultiAddr(str string) multiaddr.Multiaddr {
	return multiaddr.StringCast(str)
}

// StringToPeerID ...
func StringToPeerID(str string) (peer.ID, error) {
	return peer.IDB58Decode(str)
}

// PeerIDToString ...
func PeerIDToString(id peer.ID) string {
	return peer.IDB58Encode(id)
}
