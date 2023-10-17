package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatRequestMsg(t *testing.T) {
	msg := FormatRequestMsg(4, 567, 4321)
	expected := &Message{
		ID: MsgRequest,
		Payload: []byte{
			0x00, 0x00, 0x00, 0x04,
			0x00, 0x00, 0x02, 0x37,
			0x00, 0x00, 0x10, 0xe1,
		},
	}
	assert.Equal(t, expected, msg)
}

func TestFormatHaveMsg(t *testing.T) {
	msg := FormatHaveMsg(1)
	expected := &Message{
		ID:      MsgHave,
		Payload: []byte{0x00, 0x00, 0x00, 0x01},
	}
	assert.Equal(t, expected, msg)
}

func TestParseHaveMsg(t *testing.T) {
	tests := map[string]struct {
		input  *Message
		output int
		fails  bool
	}{
		"correct input": {
			input:  &Message{ID: MsgHave, Payload: []byte{0x00, 0x00, 0x00, 0x01}},
			output: 1,
			fails:  false,
		},
		"invalid message type": {
			input:  &Message{ID: MsgPiece, Payload: []byte{0x00, 0x00, 0x00, 0x01}},
			output: 0,
			fails:  true,
		},
		"invalid payload length: too short": {
			input:  &Message{ID: MsgHave, Payload: []byte{0x00, 0x00, 0x01}},
			output: 0,
			fails:  true,
		},
		"invalid payload length: too long": {
			input:  &Message{ID: MsgHave, Payload: []byte{0x00, 0x00, 0x00, 0x00, 0x01}},
			output: 0,
			fails:  true,
		},
	}

	for _, test := range tests {
		index, err := ParseHaveMsg(test.input)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, index)
	}
}

func TestParsePieceMsg(t *testing.T) {
	tests := map[string]struct {
		inputIndex int
		inputBuf   []byte
		inputMsg   *Message
		outputN    int
		outputBuf  []byte
		fails      bool
	}{
		"correct input": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04,
					0x00, 0x00, 0x00, 0x02,
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
				},
			},
			outputBuf: []byte{0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x00},
			outputN:   6,
			fails:     false,
		},
		"invalid message type": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID:      MsgChoke,
				Payload: []byte{},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"invalid index": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x05, // <- fails here
					0x00, 0x00, 0x00, 0x02,
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"invalid offset": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04,
					0x00, 0x00, 0x00, 0x0c, // <-- fails here
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"invalid payload length: too long": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04,
					0x00, 0x00, 0x00, 0x02,
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x0a, 0x0b, 0x0c, 0x0d, // <-- fails here
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"invalid payload length: too short": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04,
					0x00, 0x00, 0x00, // <-- fails here
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
	}

	for _, test := range tests {
		n, err := ParsePieceMsg(test.inputIndex, test.inputBuf, test.inputMsg)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.outputBuf, test.inputBuf)
		assert.Equal(t, test.outputN, n)
	}
}

func TestReadMessage(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output *Message
		fails  bool
	}{
		"correct input": {
			input:  []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
			output: &Message{ID: MsgHave, Payload: []byte{1, 2, 3, 4}},
			fails:  false,
		},
		"keep-alive message": {
			input:  []byte{0, 0, 0, 0},
			output: &Message{ID: MsgKeepAlive, Payload: nil},
			fails:  false,
		},
		"invalid length: too short": {
			input:  []byte{1, 2, 3},
			output: nil,
			fails:  true,
		},
		"invalid length: too long for buffer": {
			input:  []byte{0, 0, 0, 5, 4, 1, 2},
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		reader := bytes.NewReader(test.input)
		m, err := ReadMessage(reader)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, m)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input  *Message
		output string
	}{
		{&Message{MsgChoke, []byte{1, 2, 3}}, "Choke: 3"},
		{&Message{MsgUnchoke, []byte{1, 2, 3}}, "Unchoke: 3"},
		{&Message{MsgInterested, []byte{1, 2, 3}}, "Interested: 3"},
		{&Message{MsgNotInterested, []byte{1, 2, 3}}, "NotInterested: 3"},
		{&Message{MsgHave, []byte{1, 2, 3}}, "Have: 3"},
		{&Message{MsgBitfield, []byte{1, 2, 3}}, "Bitfield: 3"},
		{&Message{MsgRequest, []byte{1, 2, 3}}, "Request: 3"},
		{&Message{MsgPiece, []byte{1, 2, 3}}, "Piece: 3"},
		{&Message{MsgCancel, []byte{1, 2, 3}}, "Cancel: 3"},
		{&Message{10, []byte{1, 2, 3}}, "!10: 3"},
	}

	for _, test := range tests {
		s := test.input.String()
		assert.Equal(t, test.output, s)
	}
}
