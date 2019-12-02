package msglist

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/asdine/storm/v3"
)

// test various abilities of CRDT if all passed with desired results, it can work properly.
// need programmatic assertion later

// for debug
func printAllPeerMsgs(p *PeerMessageList) {
	latestID := p.LatestMsgID
	var messageIDs []int
	for i := 1; i <= latestID; i++ {
		messageIDs = append(messageIDs, i)
	}
	messages, _ := p.FindMessage(messageIDs)

	log.Printf("--> going to print all messages of %s with lastest peer %d", p.PeerName, p.LatestMsgID)
	for _, message := range messages {
		log.Printf("%d (author:%s, sender:%s) content:%s", message.ID, message.AuthorID, message.SenderID, message.Content)
	}
}

func TestNew(t *testing.T) {
	os.Remove("./test.db")
	pid1 := "tom"
	pid2 := "jerry"
	db, err := storm.Open(fmt.Sprintf("./test.db"))
	if err != nil {
		log.Println(err)
	}

	ml1, err := New(pid1, "littleT", []string{"1", "2"}, db.From(pid1))
	ml2, err := New(pid2, "bigJ", []string{"3", "4"}, db.From(pid2))
	printAllPeerMsgs(ml1)
	printAllPeerMsgs(ml2)
	db.Close()
}

func TestAdd(t *testing.T) {
	pid1 := "tom"
	pid2 := "jerry"
	db, err := storm.Open(fmt.Sprintf("./test.db"))
	if err != nil {
		log.Println(err)
	}

	ml1, err := New(pid1, "littleT", []string{"1", "2"}, db.From(pid1))
	ml2, err := New(pid2, "bigJ", []string{"3", "4"}, db.From(pid2))
	msg1 := Message{
		ID:        1,
		SenderID:  pid1,
		AuthorID:  pid1,
		Content:   "hello world!",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg1)
	msg2 := Message{
		ID:        2,
		SenderID:  pid1,
		AuthorID:  pid1,
		Content:   "anyone?",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg2)
	printAllPeerMsgs(ml1)
	printAllPeerMsgs(ml2)
	db.Close()
}

func TestAddRepeat(t *testing.T) {
	pid1 := "tom"
	pid2 := "jerry"
	db, err := storm.Open(fmt.Sprintf("./test.db"))
	if err != nil {
		log.Println(err)
	}

	ml1, err := New(pid1, "littleT", []string{"1", "2"}, db.From(pid1))
	msg1 := Message{
		ID:        1,
		SenderID:  pid2,
		AuthorID:  pid1,
		Content:   "hello world?????",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg1)
	msg2 := Message{
		ID:        2,
		SenderID:  pid1,
		AuthorID:  pid1,
		Content:   "anyone!!!!!!!",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg2)
	printAllPeerMsgs(ml1)
	db.Close()
}

func TestAddReplace(t *testing.T) {
	pid1 := "tom"
	pid2 := "jerry"
	db, err := storm.Open(fmt.Sprintf("./test.db"))
	if err != nil {
		log.Println(err)
	}

	ml1, err := New(pid1, "littleT", []string{"1", "2"}, db.From(pid1))
	msg1 := Message{
		ID:        5,
		SenderID:  pid2,
		AuthorID:  pid1,
		Content:   "wait???",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg1)
	msg2 := Message{
		ID:        6,
		SenderID:  pid2,
		AuthorID:  pid1,
		Content:   "Yes???",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg2)
	printAllPeerMsgs(ml1)
	msg1 = Message{
		ID:        5,
		SenderID:  pid1,
		AuthorID:  pid1,
		Content:   "wait!!!",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg1)
	msg2 = Message{
		ID:        6,
		SenderID:  pid1,
		AuthorID:  pid1,
		Content:   "Yes!!!",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg2)
	printAllPeerMsgs(ml1)

	db.Close()
}

func TestAddBackward(t *testing.T) {
	pid1 := "tom"
	pid2 := "jerry"
	db, err := storm.Open(fmt.Sprintf("./test.db"))
	if err != nil {
		log.Println(err)
	}

	ml1, err := New(pid1, "littleT", []string{"1", "2"}, db.From(pid1))
	msg1 := Message{
		ID:        3,
		SenderID:  pid2,
		AuthorID:  pid1,
		Content:   "Back???",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg1)
	msg2 := Message{
		ID:        4,
		SenderID:  pid1,
		AuthorID:  pid1,
		Content:   "For What???",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg2)
	printAllPeerMsgs(ml1)
	db.Close()
}

func TestAddMerge(t *testing.T) {
	pid1 := "tom"
	pid2 := "jerry"
	os.Remove("./test.db")
	db, err := storm.Open(fmt.Sprintf("./test.db"))
	if err != nil {
		log.Println(err)
	}

	ml1, err := New(pid1, "littleT", []string{"1", "2"}, db.From(pid1))
	msg1 := Message{
		ID:        2,
		SenderID:  pid2,
		AuthorID:  pid1,
		Content:   "Back???",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg1)
	msg2 := Message{
		ID:        3,
		SenderID:  pid2,
		AuthorID:  pid1,
		Content:   "For What???",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg2)
	printAllPeerMsgs(ml1)
	msg1 = Message{
		ID:        1,
		SenderID:  pid2,
		AuthorID:  pid1,
		Content:   "Hello???",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg1)
	msg2 = Message{
		ID:        2,
		SenderID:  pid1,
		AuthorID:  pid1,
		Content:   "Front???",
		CreatedAt: time.Now().Unix(),
	}
	ml1.Add(msg2)
	printAllPeerMsgs(ml1)
	db.Close()
}

func TestSave(t *testing.T) {
	pid1 := "tom"
	db, err := storm.Open(fmt.Sprintf("./test.db"))
	if err != nil {
		log.Println(err)
	}

	ml1, err := New(pid1, "littleT", []string{"1", "2"}, db.From(pid1))
	ml1.PeerName = "bigT"
	ml1.Save()

	log.Println(ml1.GetPeerData().MsgDict)

	db.Close()
	os.Remove("./test.db")
}
