package network

import (
	"fmt"
	"testing"

	keygen "github.com/cyber-rhizome/keygen"
)

func TestCreateHost(t *testing.T) {
	protString, _ := keygen.CreateProtString()
	prot, _ := keygen.StringToProt(protString)
	privKey, _ := keygen.CreatePrivKey("test", protString)
	h, err := CreateHost(privKey, prot, 8080)
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}
	fmt.Println(h.ID())
	fmt.Println(h.Addrs())
	err = h.Close()
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}
}

func TestCreatePubSub(t *testing.T) {
	protString, _ := keygen.CreateProtString()
	prot, _ := keygen.StringToProt(protString)
	privKey, _ := keygen.CreatePrivKey("test", protString)
	h, err := CreateHost(privKey, prot, 8080)
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}
	fmt.Println(h.ID())
	fmt.Println(h.Addrs())
	err = h.Close()
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}
	pubsub, err := CreatePubSub(h)
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}
	pubsub.Subscribe("test")
	fmt.Println(pubsub.GetTopics())
}
