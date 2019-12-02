package utility

import (
	"testing"

	"github.com/multiformats/go-multiaddr"
)

func TestUtility(t *testing.T) {
	m := StringToMultiAddr("/ip4/127.0.0.1/tcp/8080/p2p/12D3KooWMPCNzj6GvGVf7T5Vtf2wwJ4M1vf56m3D46zUpFzEQgCi")
	t.Log(m)
	s := MultiaddrToString(m)
	t.Log(s)
	id, err := StringToPeerID("12D3KooWMPCNzj6GvGVf7T5Vtf2wwJ4M1vf56m3D46zUpFzEQgCi")
	if err != nil {
		t.Error(err)
	}
	t.Log(id)
	sid := PeerIDToString(id)
	t.Log(sid)
	pTom := SplitP2PAddrs([]multiaddr.Multiaddr{m})
	t.Log(pTom)
	p2pm := BuildP2PAddrsFromString("12D3KooWMPCNzj6GvGVf7T5Vtf2wwJ4M1vf56m3D46zUpFzEQgCi", pTom["12D3KooWMPCNzj6GvGVf7T5Vtf2wwJ4M1vf56m3D46zUpFzEQgCi"])
	t.Log(p2pm)
}
