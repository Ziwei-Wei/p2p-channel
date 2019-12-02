package main

import (
	"fmt"
	"log"

	"github.com/Ziwei-Wei/cyber-rhizome-host/channel"
	"github.com/Ziwei-Wei/cyber-rhizome-host/keygen"
	"github.com/Ziwei-Wei/cyber-rhizome-host/network"
	"github.com/asdine/storm/v3"
)

func main() {
	protString, _ := keygen.CreateProtString()
	prot, _ := keygen.StringToProt(protString)
	privKey, _ := keygen.CreatePrivKey("test", protString)
	h, err := network.CreateHost(privKey, prot, 8080)
	if err != nil {
		fmt.Println(err)
	}
	err = h.Close()
	if err != nil {
		fmt.Println(err)
	}
	pubsub, err := network.CreatePubSub(h)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pubsub.GetTopics())
	db, err := storm.Open("./db/test.db")
	if err != nil {
		fmt.Println(err)
	}

	err = channel.CreateChannel("tester1", "testChannel", "chat", db, h, pubsub)
	if err != nil {
		fmt.Println(err)
	}

	var c channel.Channel
	c, err = channel.OpenChannel("tester1", "testChannel", "chat", db, h, pubsub)
	if err != nil {
		fmt.Println(err)
	}
	c.GetPeers()
	c.On("PeerJoin", func(peer interface{}) {
		log.Println("hello")
		fmt.Println(peer)
	})
	for {

	}
}
