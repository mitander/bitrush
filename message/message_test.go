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
