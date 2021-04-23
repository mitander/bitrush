package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

type Message struct {
	ID      messageID
	Payload []byte
}

// [https://wiki.theory.org/BitTorrentSpecification#Messages]
const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

func FormatRequestMsg(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{
		ID:      MsgRequest,
		Payload: payload,
	}
}

func FormatHaveMsg(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{
		ID:      MsgHave,
		Payload: payload,
	}
}

func ParseHaveMsg(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("parsing have message failed: message id (id:%d - expected: %d)", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("parsing have message failed: payload length (input: %d - expected: 4)", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

func ParsePieceMsg(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("parsing piece message failed: message id (id:%d - expected:%d)", MsgPiece, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("parsing piece message failed: payload length - (id: %d - expected: 8)", len(msg.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("parsing piece message failed: invalid index (index: %d - expected: %d)", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("parsing piece message failed: invalid buffer length (begin:%d - expected: %d)", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("parsing piece message failed: invalid data length (data:%d - expected: %d)", data, len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}

func (msg *Message) Serialize() []byte {
	if msg == nil {
		return make([]byte, 4)
	}
	length := uint32(len(msg.Payload) + 1) // Payload + messageID
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(msg.ID)
	copy(buf[5:], msg.Payload)
	return buf
}

func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, nil
	}
	msgBuf := make([]byte, length)
	_, err = io.ReadFull(r, msgBuf)
	if err != nil {
		return nil, err
	}
	return &Message{
		ID:      messageID(msgBuf[0]),
		Payload: msgBuf[1:],
	}, nil
}

func (msg *Message) name() string {
	if msg.Payload == nil {
		return "KeepAlive"
	}
	switch msg.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("!%d", msg.ID)
	}
}

func (msg *Message) String() string {
	if msg == nil {
		return msg.name()
	}
	return fmt.Sprintf("%s: %d", msg.name(), len(msg.Payload))
}
