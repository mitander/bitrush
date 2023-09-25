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
		log.Error("failed to set piece: invalid byte index")
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
		err := errors.New("invalid bitfield received: msg is nil")
		log.Error(err.Error())
		return nil, err
	}

	if msg.ID != message.MsgBitfield {
		err := errors.New("invalid bitfield received: wrong id")
		log.WithFields(log.Fields{"got": msg.ID, "expected": message.MsgBitfield}).Error(err.Error())
		return nil, err
	}

	return msg.Payload, nil
}
