package p2p

import (
	"errors"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type Bitfield []byte

func (b Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(b) {
		log.Warnf("failed to set piece: invalid byte index: %d", byteIndex)
		return false
	}
	return b[byteIndex]>>uint(7-offset)&1 != 0
}

func (b Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	if byteIndex < 0 || byteIndex >= len(b) {
		log.Warnf("failed to set piece: invalid byte index: %d", byteIndex)
		return
	}
	b[byteIndex] |= 1 << uint(7-offset)
}

func RecvBitfield(conn net.Conn) (Bitfield, error) {
	// set deadline to fail instead of blocking after 5 seconds
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := ReadMessage(conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, errors.New("invalid bitfield received: msg is nil")
	}

	if msg.ID != MsgBitfield {
		return nil, errors.New("invalid bitfield received: wrong id")
	}

	return msg.Payload, nil
}
