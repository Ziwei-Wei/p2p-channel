package keygen

import (
	"fmt"
	"testing"
)

func TestProtectorKey(t *testing.T) {
	protString, err := CreateProtString()
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}
	t.Log(protString)

	protKey, err := StringToProt(protString)
	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}
	t.Log(protKey)
}

func TestPrivateKey(t *testing.T) {
	userName1 := "tester"
	password1 := "12345678"

	userName2 := "tester"
	password2 := "123456789"

	key1, err := CreatePrivKey(userName1, password1)
	key1Copy, err := CreatePrivKey(userName1, password1)

	key2, err := CreatePrivKey(userName2, password2)

	if err != nil {
		t.Errorf(fmt.Sprintln(err))
	}
	if key1.Equals(key2) == true {
		t.Errorf(fmt.Sprintln("should not produce same key"))
	}
	if key1.Equals(key1Copy) == false {
		t.Errorf(fmt.Sprintln("should produce same key"))
	}
}
