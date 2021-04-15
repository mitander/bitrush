package bitfield

import (
	"fmt"
	"net"
	"time"

	"github.com/mitander/bitrush/message"
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
		err := fmt.Errorf("recieving bitfield failed: msg is nil [id: %d]", msg.ID)
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("recieving bitfield failed: wrong msg id [id: %d] - [expected: 5]", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}
