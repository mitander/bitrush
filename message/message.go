package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

var (
	InvalidMessageId     error = errors.New("invalid message id")
	InvalidPayloadLength error = errors.New("invalid payload length")
	InvalidMessageIndex  error = errors.New("invalid payload length")
	InvalidBufferLength  error = errors.New("invalid buffer length")
	InvalidDataLength    error = errors.New("invalid data length")
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
		log.WithFields(log.Fields{"got": msg.ID, "expected": MsgHave}).Error(InvalidMessageId.Error())
		return 0, InvalidMessageId
	}

	if len(msg.Payload) != 4 {
		log.WithFields(log.Fields{"got": len(msg.Payload), "expected": 4}).Error(InvalidPayloadLength.Error())
		return 0, InvalidPayloadLength
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

func ParsePieceMsg(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		log.WithFields(log.Fields{"got": msg.ID, "expected": MsgPiece}).Error(InvalidMessageId.Error())
		return 0, InvalidMessageId
	}

	if len(msg.Payload) < 8 {
		log.WithFields(log.Fields{"got": len(msg.Payload), "expected": 8}).Error(InvalidPayloadLength.Error())
		return 0, InvalidPayloadLength
	}

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		log.WithFields(log.Fields{"got": parsedIndex, "expected": index}).Error(InvalidMessageIndex.Error())
		return 0, InvalidMessageIndex
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		log.WithFields(log.Fields{"got": begin, "expected-over": len(buf)}).Error(InvalidBufferLength.Error())
		return 0, InvalidBufferLength
	}

	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		log.WithFields(log.Fields{"got": begin + len(data), "expected-over": len(buf)}).Error(InvalidDataLength.Error())
		return 0, InvalidDataLength
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
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed reading to buffer")
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		err := errors.New("length cannot be 0")
		log.Error(err.Error())
		return nil, err
	}
	msgBuf := make([]byte, length)
	_, err = io.ReadFull(r, msgBuf)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed reading to buffer")
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
