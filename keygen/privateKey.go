package keygen

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	"github.com/libp2p/go-libp2p-core/peer"

	crypto "github.com/libp2p/go-libp2p-core/crypto"
)

// CreatePrivKey create a unique private key, hash username and protector key json to a [32]byte
func CreatePrivKey(userName string, protector string) (crypto.PrivKey, error) {
	account := account{
		UserName: userName,
		ProtKey:  protector,
	}

	data, _ := json.Marshal(&account)
	hasher := sha256.New()
	hasher.Write(data)
	d := hasher.Sum(nil)

	reader := bytes.NewReader(d)
	PrivKey, _, err := crypto.GenerateEd25519Key(reader)
	if err != nil {
		return nil, err
	}
	return PrivKey, nil
}

// PrivKeyToPeerID convert private key to PeerID
func PrivKeyToPeerID(PrivKey crypto.PrivKey) (peer.ID, error) {
	return peer.IDFromPublicKey(PrivKey.GetPublic())
}
