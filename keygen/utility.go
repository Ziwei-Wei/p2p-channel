package keygen

import (
	pnet "github.com/libp2p/go-libp2p-pnet"
)

type account struct {
	UserName string
	ProtKey  string
}

// create NetProtector bytes
func createProtBytes() ([]byte, error) {
	bytes, err := pnet.GenerateV1Bytes()
	if err != nil {
		return nil, err
	}
	return (*bytes)[:], nil
}

// create NetProtector string
func protStringToBytes(protString string) [32]byte {
	binaryProt := [32]byte{}
	copy(binaryProt[:], protString)
	return binaryProt
}
