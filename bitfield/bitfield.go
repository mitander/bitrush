package bitfield

import (
	"errors"
	"net"
	"time"

	"github.com/mitander/bitrush/message"
	log "github.com/sirupsen/logrus"
)

type Bitfield []byte

func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}
	return bf[byteIndex]>>uint(7-offset)&1 != 0
}

func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	if byteIndex < 0 || byteIndex >= len(bf) {
		log.Warnf("failed to set piece: invalid byte index: %d", byteIndex)
		return
	}
	bf[byteIndex] |= 1 << uint(7-offset)
}

func RecvBitfield(conn net.Conn) (Bitfield, error) {
	// set deadline to fail instead of blocking after 5 seconds
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, errors.New("invalid bitfield received: msg is nil")
	}

	if msg.ID != message.MsgBitfield {
		return nil, errors.New("invalid bitfield received: wrong id")
	}

	return msg.Payload, nil
}
