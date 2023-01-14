package core

import (
	"github.com/wisepythagoras/owncom/crypto"
)

const PACKET_SIZE = 100

type Message struct {
	Msg []byte
}

func (m *Message) PacketsAESGCM(key, salt []byte) ([]Packet, error) {
	key, err := crypto.PBKDF2Key(key, salt)

	if err != nil {
		return nil, err
	}

	ciphertext, err := crypto.EncryptGCM(m.Msg, key)

	if err != nil {
		return nil, err
	}

	packets := make([]Packet, 0)
	numOfPackets := len(ciphertext) / PACKET_SIZE
	remainderBytes := len(ciphertext) % PACKET_SIZE

	for i := 0; i < numOfPackets; i++ {
		packet := Packet{
			Content: ciphertext[PACKET_SIZE*i : PACKET_SIZE*(i+1)],
			ID:      uint32(numOfPackets) + 1,
		}
		packets = append(packets, packet)
	}

	if remainderBytes > 0 {
		packet := Packet{
			Content: ciphertext[len(ciphertext)-remainderBytes:],
			ID:      uint32(numOfPackets) + 1,
		}
		packets = append(packets, packet)
	}

	return packets, nil
}

func UnmarshalAESEncryptedMessage(key, salt []byte) (*Message, error) {
	return nil, nil
}
