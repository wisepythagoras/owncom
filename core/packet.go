package core

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"io"

	"github.com/vmihailenco/msgpack"
)

type Packet struct {
	Content  []byte `msgpack:"d"`
	Checksum uint16 `msgpack:"c"`
}

func (p *Packet) Marshal() ([]byte, error) {
	p.Checksum = p.GetChecksum()

	var buff bytes.Buffer
	var err error

	raw, err := msgpack.Marshal(p)

	if err != nil {
		return nil, err
	}

	writer := gzip.NewWriter(&buff)

	if _, err = writer.Write(raw); err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (p *Packet) MarshalToHex() (string, error) {
	buff, err := p.Marshal()

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buff), nil
}

func (p *Packet) GetChecksum() uint16 {
	return bytesChecksum(p.Content)
}

func UnmarshalPacket(b []byte) (*Packet, error) {
	var buff bytes.Buffer
	var err error
	inBuff := bytes.NewBuffer(b)

	reader, err := gzip.NewReader(inBuff)

	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(&buff, reader); err != nil {
		return nil, err
	}

	if err = reader.Close(); err != nil {
		return nil, err
	}

	packet := new(Packet)

	if err := msgpack.Unmarshal(buff.Bytes(), &packet); err != nil {
		return nil, err
	}

	return packet, nil
}
