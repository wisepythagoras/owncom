package core

import (
	"github.com/wisepythagoras/owncom/crypto"
)

type Message struct {
	Msg []byte
}

func (m *Message) MarshalAESEncrypted(key, salt []byte) ([]byte, error) {
	key, err := crypto.PBKDF2Key(key, salt)

	if err != nil {
		return nil, err
	}

	ciphertext, err := crypto.EncryptGCM(m.Msg, key)

	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

func UnmarshalAESEncryptedMessage(key, salt []byte) (*Message, error) {
	return nil, nil
}
