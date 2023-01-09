package core

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"io"
)

type Packet struct {
	Content string
}

func (p *Packet) Marshal() ([]byte, error) {
	var buff bytes.Buffer
	var err error

	writer := gzip.NewWriter(&buff)

	if _, err = writer.Write([]byte(p.Content)); err != nil {
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

	packet := &Packet{Content: buff.String()}

	return packet, nil
}
