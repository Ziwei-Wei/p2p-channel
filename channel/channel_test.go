package channel

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/asdine/storm/v3"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/cyber-rhizome/keygen"
	"github.com/cyber-rhizome/network"
)

// func TestTwoPeerInOneLocalChannel(t *testing.T) {
// 	os.Remove("./test.db")
// 	os.Remove("./test.db.lock")
// 	protString, _ := keygen.CreateProtString()
// 	protector, _ := keygen.StringToProt(protString)
// 	u0 := "tester0"
// 	u1 := "tester1"
// 	cName := "testChannel"

// 	db, err := storm.Open(fmt.Sprintf("./%s.db", "test"))
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	err = CreateChannel(6666, cName, "chat", u0, protString, db)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}
// 	err = CreateChannel(6667, cName, "chat", u1, protString, db)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}

// 	privKey0, _ := keygen.CreatePrivKey(u0, protString)
// 	h0, _ := network.CreateHost(privKey0, protector, 6666)
// 	p0, _ := network.CreatePubSub(h0)

// 	privKey1, _ := keygen.CreatePrivKey(u1, protString)
// 	h1, _ := network.CreateHost(privKey1, protector, 6667)
// 	p1, _ := network.CreatePubSub(h1)

// 	c0, err := OpenChannel(context.Background(), cName, "chat", h0, p0, db)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}

// 	c1, err := OpenChannel(context.Background(), cName, "chat", h1, p1, db)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}
// 	c0.On("NewMessage", func(c *ChatChannel, data interface{}) {
// 		log.Printf("test-> OnNewMessage %v from %v", data.(chatMessage).Content, data.(chatMessage).AuthorID)
// 		list := c.peerIDToMsgList[c.userID]
// 		log.Printf("current messages in %v", c.userID)
// 		msgs, _ := list.FindMessage([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
// 		for _, msg := range msgs {
// 			log.Println(msg.Content)
// 		}
// 	})
// 	for i := 0; i < 4; i++ {
// 		c0.Send(fmt.Sprintf("%d", i))
// 		c1.Send(fmt.Sprintf("%d", i))
// 	}
// 	time.Sleep(5 * time.Second)
// 	c0.Leave()
// 	c1.Leave()
// 	db.Close()
// 	h0.Close()
// 	h1.Close()
// 	os.Remove("./test.db")
// 	os.Remove("./test.db.lock")
// 	time.Sleep(10 * time.Second)
// }

// func TestTwoPeerSynchronizationTask(t *testing.T) {
// 	os.Remove("./test.db")
// 	os.Remove("./test0.db")
// 	os.Remove("./test0.db.lock")
// 	os.Remove("./test1.db")
// 	os.Remove("./test1.db.lock")

// 	protString, _ := keygen.CreateProtString()
// 	protector, _ := keygen.StringToProt(protString)
// 	u0 := "tester0"
// 	u1 := "tester1"
// 	cName := "testChannel"

// 	db0, err := storm.Open(fmt.Sprintf("./%s0.db", "test"))
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	db1, err := storm.Open(fmt.Sprintf("./%s1.db", "test"))
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	err = CreateChannel(6666, cName, "chat", u0, protString, db0)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}
// 	err = CreateChannel(6667, cName, "chat", u1, protString, db1)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}

// 	privKey0, _ := keygen.CreatePrivKey(u0, protString)
// 	h0, _ := network.CreateHost(privKey0, protector, 6666)
// 	p0, _ := network.CreatePubSub(h0)

// 	privKey1, _ := keygen.CreatePrivKey(u1, protString)
// 	h1, _ := network.CreateHost(privKey1, protector, 6667)
// 	p1, _ := network.CreatePubSub(h1)

// 	c0, err := OpenChannel(context.Background(), cName, "chat", h0, p0, db0)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}

// 	log.Println("--> you need to accept for network access!")
// 	time.Sleep(5 * time.Second)

// 	log.Printf("tester0 will send 10 messages")
// 	for i := 0; i < 10; i++ {
// 		log.Printf("tester0 sent message %d ", i)
// 		c0.Send(fmt.Sprintf("%d", i))
// 	}

// 	log.Printf("tester1 will login after 2 seconds")
// 	time.Sleep(2 * time.Second)

// 	c1, err := OpenChannel(context.Background(), cName, "chat", h1, p1, db1)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}

// 	c1.On("NewPeer", func(c *ChatChannel, data interface{}) {
// 		list := c.peerIDToMsgList[c0.userID]
// 		msgs, _ := list.FindMessage([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
// 		log.Printf("When tester1 started, it have %v tester0 message...", len(msgs))
// 		for _, msg := range msgs {
// 			log.Printf("Author:%s, Content:%s, CreatedAt:%d", c.peerIDToMsgList[msg.AuthorID].PeerName, msg.Content, msg.CreatedAt)
// 		}
// 	})

// 	finished := false
// 	c1.On("Sync", func(c *ChatChannel, data interface{}) {
// 		list := c.peerIDToMsgList[c0.userID]
// 		msgs, _ := list.FindMessage([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
// 		log.Printf("Aftter sync, it have %v tester0 message...", len(msgs))
// 		for _, msg := range msgs {
// 			log.Printf("Author:%s, Content:%s, CreatedAt:%d", c.peerIDToMsgList[msg.AuthorID].PeerName, msg.Content, msg.CreatedAt)
// 		}
// 		finished = true
// 	})

// 	c1.AddPeer(c0.userID, u0, c0.peerIDToMsgList[c0.userID].P2PAddrs)

// 	log.Println("Wait for magic...(peers wil seek to sync with 2 nearest peers and the target if it detect it lacks message, every n second)")
// 	for finished == false {
// 		time.Sleep(time.Second)
// 	}
// 	log.Println("Going to close...")

// 	c0.Leave()
// 	c1.Leave()
// 	db0.Close()
// 	db1.Close()
// 	h0.Close()
// 	h1.Close()
// 	os.Remove("./test.db")
// 	os.Remove("./test0.db")
// 	os.Remove("./test0.db.lock")
// 	os.Remove("./test1.db")
// 	os.Remove("./test1.db.lock")
// 	time.Sleep(10 * time.Second)
// }

// func TestTwoPeerSyncStressTesting(t *testing.T) {
// 	os.Remove("./test.db")
// 	os.Remove("./test0.db")
// 	os.Remove("./test0.db.lock")
// 	os.Remove("./test1.db")
// 	os.Remove("./test1.db.lock")

// 	protString, _ := keygen.CreateProtString()
// 	protector, _ := keygen.StringToProt(protString)
// 	u0 := "tester0"
// 	u1 := "tester1"
// 	cName := "testChannel"

// 	db0, err := storm.Open(fmt.Sprintf("./%s0.db", "test"))
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	db1, err := storm.Open(fmt.Sprintf("./%s1.db", "test"))
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	err = CreateChannel(6666, cName, "chat", u0, protString, db0)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}
// 	err = CreateChannel(6667, cName, "chat", u1, protString, db1)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}

// 	privKey0, _ := keygen.CreatePrivKey(u0, protString)
// 	h0, _ := network.CreateHost(privKey0, protector, 6666)
// 	p0, _ := network.CreatePubSub(h0)

// 	privKey1, _ := keygen.CreatePrivKey(u1, protString)
// 	h1, _ := network.CreateHost(privKey1, protector, 6667)
// 	p1, _ := network.CreatePubSub(h1)

// 	c0, err := OpenChannel(context.Background(), cName, "chat", h0, p0, db0)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}

// 	log.Println("--> you need to accept for network access!")
// 	time.Sleep(5 * time.Second)

// 	log.Printf("tester0 will send 1000 messages, will print 1 message every 100 messages")
// 	var msgL []int
// 	for i := 1; i <= 1000; i++ {
// 		if i%100 == 0 {
// 			log.Printf("tester0 sent message %d ", i)
// 		}
// 		msgL = append(msgL, i)
// 		c0.Send(fmt.Sprintf("%d", i))
// 	}

// 	log.Printf("tester1 will login after 2 seconds")
// 	time.Sleep(2 * time.Second)

// 	c1, err := OpenChannel(context.Background(), cName, "chat", h1, p1, db1)
// 	if err != nil {
// 		t.Errorf(fmt.Sprintln(err))
// 	}

// 	c1.On("NewPeer", func(c *ChatChannel, data interface{}) {
// 		list := c.peerIDToMsgList[c0.userID]
// 		msgs, _ := list.FindMessage(msgL)
// 		log.Printf("When tester1 started, it have %v tester0 message...", len(msgs))
// 		for i, msg := range msgs {
// 			if (i+1)%100 == 0 {
// 				log.Printf("ID:%d, Author:%s, Content:%s, CreatedAt:%d", msg.ID, c.peerIDToMsgList[msg.AuthorID].PeerName, msg.Content, msg.CreatedAt)
// 			}
// 		}
// 	})

// 	finished := false
// 	c1.On("Sync", func(c *ChatChannel, data interface{}) {
// 		list := c.peerIDToMsgList[c0.userID]
// 		msgs, _ := list.FindMessage(msgL)
// 		log.Printf("Aftter sync, it have %v tester0 message...", len(msgs))
// 		log.Printf("will print 1 message every 1000 messages")
// 		for i, msg := range msgs {
// 			if (i+1)%100 == 0 {
// 				log.Printf("ID:%d, Author:%s, Content:%s, CreatedAt:%d", msg.ID, c.peerIDToMsgList[msg.AuthorID].PeerName, msg.Content, msg.CreatedAt)
// 			}
// 		}
// 		finished = true
// 	})

// 	c1.AddPeer(c0.userID, u0, c0.peerIDToMsgList[c0.userID].P2PAddrs)

// 	log.Println("Wait for magic...(peers wil seek to sync with 2 nearest peers and the target if it detect it lacks message, every n second)")

// 	for finished == false {
// 		time.Sleep(10 * time.Second)
// 	}
// 	log.Println("Going to close...")

// 	c0.Leave()
// 	c1.Leave()
// 	db0.Close()
// 	db1.Close()
// 	h0.Close()
// 	h1.Close()
// 	os.Remove("./test.db")
// 	os.Remove("./test0.db")
// 	os.Remove("./test0.db.lock")
// 	os.Remove("./test1.db")
// 	os.Remove("./test1.db.lock")
// 	time.Sleep(10 * time.Second)
// }

// test with 40 peers (400 connections, 400 messages in a few seconds)
// if we run too much peer on same machine soon the bottle neck will be cpu.
func TestTooMuchPeersAndMessageStressTesting(t *testing.T) {
	peerAmount := 40
	msgAmount := 10
	connectionLimit := 10

	protString, _ := keygen.CreateProtString()
	protector, _ := keygen.StringToProt(protString)

	msgL := make([]int, msgAmount)
	user := make([]string, peerAmount)
	db := make([]*storm.DB, peerAmount)
	key := make([]crypto.PrivKey, peerAmount)
	h := make([]host.Host, peerAmount)
	p := make([]*pubsub.PubSub, peerAmount)
	c := make([]*ChatChannel, peerAmount)
	cName := "testChannel"
	log.Println("====> you need to accept for network access!")
	for i := 0; i < msgAmount; i++ {
		msgL[i] = i + 1
	}
	log.Println("====> load test")
	for i := 0; i < peerAmount; i++ {
		os.Remove(fmt.Sprintf("./test%d.db", i))
		os.Remove(fmt.Sprintf("./test%d.db.lock", i))
		user[i] = fmt.Sprintf("tester%d", i)
		currDB, err := storm.Open(fmt.Sprintf("./test%d.db", i))
		if err != nil {
			log.Println(err)
			return
		}
		db[i] = currDB
		err = CreateChannel(10000+i, cName, "chat", user[i], protString, db[i])
		if err != nil {
			t.Errorf(fmt.Sprintln(err))
		}
		currKey, _ := keygen.CreatePrivKey(user[i], protString)
		key[i] = currKey
		currH, _ := network.CreateHost(key[i], protector, uint16(10000+i))
		h[i] = currH
		currP, _ := network.CreatePubSub(h[i])
		p[i] = currP
		currC, err := OpenChannel(context.Background(), cName, "chat", h[i], p[i], db[i])
		if err != nil {
			t.Errorf(fmt.Sprintln(err))
		}
		c[i] = currC
		c[i].On("UserLeave", func(c *ChatChannel, data interface{}) {
			if c.peerIDToMsgList[c.userID].PeerName == user[0] {
				log.Printf("====> From tester%d's point of view:", 0)
				for _, list := range c.peerIDToMsgList {
					msgs, _ := list.FindMessage(msgL)
					log.Printf("==> %d messages from peer %s latestID is %d", len(msgs), list.PeerName, list.LatestMsgID)
					// for i, msg := range msgs {
					// 	if (i+1)%1 == 0 {
					// 		log.Printf("ID:%d, Author:%s, Content:%s, CreatedAt:%d", msg.ID, c.peerIDToMsgList[msg.AuthorID].PeerName, msg.Content, msg.CreatedAt)
					// 	}
					// }
				}
			}
		})
		c[i].On("Sync", func(c *ChatChannel, data interface{}) {
			if c.peerIDToMsgList[c.userID].PeerName == user[peerAmount-1] {
				log.Printf("====> From tester%d's point of view:", peerAmount-1)
				for _, list := range c.peerIDToMsgList {
					msgs, _ := list.FindMessage(msgL)
					log.Printf("==> %d messages from peer %s latestID is %d", len(msgs), list.PeerName, list.LatestMsgID)
				}
			}
		})
	}
	log.Println("====> test loaded")
	time.Sleep(5 * time.Second)

	log.Println("====> build some connections")
	for i := 0; i < peerAmount; i++ {
		p := i
		j := i
		for j < i+connectionLimit {
			j++
			p++
			if p >= peerAmount {
				p -= peerAmount
			}
			time.Sleep(100 * time.Millisecond)
			err := c[i].AddPeer(c[p].userID, user[p], c[p].peerIDToMsgList[c[p].userID].P2PAddrs)
			if err != nil {
				log.Printf("error: %s", err)
			}
		}
	}
	log.Println("====> connections built")
	time.Sleep(5 * time.Second)

	log.Println("====> send messages")
	for i := 0; i < msgAmount; i++ {
		for j := 0; j < peerAmount; j++ {
			time.Sleep(100 * time.Millisecond)
			c[j].Send(fmt.Sprintf("%d", i+1))
		}
	}
	log.Println("====> messages sent")
	time.Sleep(50 * time.Second)

	log.Println("====> check messages")
	for i := 0; i < peerAmount; i++ {
		c[i].Leave()
		db[i].Close()
		h[i].Close()
		os.Remove(fmt.Sprintf("./test%d.db", i))
		os.Remove(fmt.Sprintf("./test%d.db.lock", i))

	}
	time.Sleep(10 * time.Second)
}
