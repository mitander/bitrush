package message

import (
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
