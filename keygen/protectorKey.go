package keygen

import (
	prot "github.com/libp2p/go-libp2p-core/pnet"
	pnet "github.com/libp2p/go-libp2p-pnet"
)

// CreateProtString create NetProtector string
func CreateProtString() (string, error) {
	bytes, err := createProtBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// StringToProt convert string to pnet.Protector
func StringToProt(protString string) (prot.Protector, error) {
	protBytes := protStringToBytes(protString)
	protKey, err := pnet.NewV1ProtectorFromBytes(&protBytes)
	return protKey, err
}
