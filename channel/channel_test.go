package channel

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/asdine/storm/v3"
	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/Ziwei-Wei/cyber-rhizome-host/keygen"
	"github.com/Ziwei-Wei/cyber-rhizome-host/network"
)

func createTestInstance(i int, protString string) (peer.ID, string, string, *storm.DB, host.Host, *pubsub.PubSub) {
	userName := fmt.Sprintf("tester%v", i)
	channelName := "testChannel"
	protector, _ := keygen.StringToProt(protString)
	privKey, _ := keygen.CreatePrivKey(userName, protString)
	pid, _ := keygen.PrivKeyToPeerID(privKey)
	host, _ := network.CreateHost(privKey, protector, 8080+uint16(i))
	pubsub, _ := network.CreatePubSub(host)
	db, err := storm.Open(fmt.Sprintf("./db/%s_%s.db", userName, channelName))
	if err != nil {
		log.Println(err)
	}
	log.Println("pid: " + pid.Pretty() + "; username: " + userName + "; channelname: " + channelName)
	return pid, userName, channelName, db, host, pubsub
}

func TestOpenChannel(t *testing.T) {
	protString, _ := keygen.CreateProtString()
	protector, _ := keygen.StringToProt(protString)
	u0 := "tester0"
	cName := "testChannel"
	db0, err := storm.Open(fmt.Sprintf("./db/%s_%s.db", "tester0", "testChannel"))
	if err != nil {
		log.Println(err)
	}

	err = CreateChannel(cName, "chat", u0, protString, db0)
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}

	privKey0, _ := keygen.CreatePrivKey(u0, protString)
	h0, _ := network.CreateHost(privKey0, protector, 8080)
	p0, _ := network.CreatePubSub(h0)

	c, err := OpenChannel(context.Background(), cName, "chat", h0, p0, db0)
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}

	c.On("PeerJoin", func(data interface{}) {
		log.Printf("test-> OnPeerJoin %s", data)
	})

	c.On("NewMessage", func(data interface{}) {
		log.Printf("test-> OnNewMessage %v", data)
	})

	c.On("Sync", func(data interface{}) {
		log.Printf("test-> OnSync %v", data)
	})

	c.On("PeerLeave", func(data interface{}) {
		log.Printf("test-> OnPeerLeave %s", data)
	})

	log.Println("peer list:")
	log.Println(c.GetPeers())
	for {
	}
}

func TestTwoPeerCommunication(t *testing.T) {

}
