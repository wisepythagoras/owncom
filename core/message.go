package core

import (
	"encoding/hex"

	"github.com/wisepythagoras/owncom/crypto"
)

const PACKET_SIZE = 50

type Message struct {
	Msg []byte
}

func (m *Message) PacketsAESGCM(aesGcmKey *crypto.AESGCMKey) ([]Packet, error) {
	key, err := crypto.PBKDF2Key(aesGcmKey.Key, aesGcmKey.Salt)

	if err != nil {
		return nil, err
	}

	ciphertext, err := crypto.EncryptGCM(m.Msg, key)

	if err != nil {
		return nil, err
	}

	hash, err := crypto.GetSHA256Hash(ciphertext)

	if err != nil {
		return nil, err
	}

	hashHex := hex.EncodeToString(hash)
	packets := make([]Packet, 0)
	numOfPackets := len(ciphertext) / PACKET_SIZE
	remainderBytes := len(ciphertext) % PACKET_SIZE

	totalPackets := numOfPackets

	if remainderBytes > 0 {
		totalPackets += 1
	}

	for i := 0; i < numOfPackets; i++ {
		packet := Packet{
			Content:   ciphertext[PACKET_SIZE*i : PACKET_SIZE*(i+1)],
			PacketNum: uint32(numOfPackets) + 1,
			Total:     uint32(totalPackets),
			ID:        hashHex[:10],
		}
		packets = append(packets, packet)
	}

	if remainderBytes > 0 {
		packet := Packet{
			Content:   ciphertext[len(ciphertext)-remainderBytes:],
			PacketNum: uint32(numOfPackets) + 1,
			Total:     uint32(totalPackets),
			ID:        hashHex[:10],
		}
		packets = append(packets, packet)
	}

	return packets, nil
}
