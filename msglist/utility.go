package msglist

import (
	"log"
)

func (p *PeerMessageList) saveNewMessageToDB(message *Message) error {
	tx, err := p.db.Begin(true)
	if err != nil {
		log.Printf("error: %v in saveNewMessageToDB, db.Begin failed", err)
		return err
	}

	defer tx.Rollback()

	var data PeerData
	err = tx.One("PeerID", message.AuthorID, &data)
	if err != nil {
		log.Printf("error: %v in saveNewMessageToDB, tx.One failed", err)
		return err
	}
	data.LatestMsgID = message.ID
	data.MsgDict[message.ID] = true

	err = tx.Save(&data)
	if err != nil {
		log.Printf("error: %v in saveNewMessageToDB, tx.Save failed", err)
		return err
	}

	err = tx.Save(message)
	if err != nil {
		log.Printf("error: %v in saveNewMessageToDB, tx.Save failed", err)
		return err
	}
	return tx.Commit()
}

func (p *PeerMessageList) saveOldMessageToDB(message *Message) error {
	tx, err := p.db.Begin(true)
	if err != nil {
		log.Printf("error: %v in saveOldMessageToDB, db.Begin failed", err)
		return err
	}
	defer tx.Rollback()

	var data PeerData
	err = tx.One("PeerID", message.AuthorID, &data)
	if err != nil {
		log.Printf("error: %v in saveOldMessageToDB, tx.One failed", err)
		return err
	}
	if data.MsgDict[message.ID] == false {
		data.MsgDict[message.ID] = true
		err = tx.Save(&data)
		if err != nil {
			log.Printf("error: %v in saveOldMessageToDB, tx.Save failed", err)
			return err
		}
		err = tx.Save(message)
		if err != nil {
			log.Printf("error: %v in saveOldMessageToDB, tx.Save failed", err)
			return err
		}

	} else {
		var oldMessage Message
		err = tx.One("ID", message.ID, &oldMessage)
		if err != nil {
			log.Printf("error: %v in saveOldMessageToDB, tx.One failed", err)
			return err
		}
		if oldMessage.SenderID != oldMessage.AuthorID {
			err = tx.Save(message)
			if err != nil {
				log.Printf("error: %v in saveOldMessageToDB, tx.Save failed", err)
				return err
			}
		}
	}
	return tx.Commit()
}
