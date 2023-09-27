package p2p

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	tests := map[string]struct {
		input  string
		output []Peer
		fails  bool
	}{
		"correct input": {
			input: string([]byte{127, 0, 2, 210, 0x1A, 0xE9, 1, 1, 1, 1, 0x00, 0x7f}),
			output: []Peer{
				{IP: net.IP{127, 0, 2, 210}, Port: 6889},
				{IP: net.IP{1, 1, 1, 1}, Port: 127},
			},
		},
		"invalid bytes in peers": {
			input:  string([]byte{127, 0, 0, 1, 0x00}),
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		peers, err := Unmarshal([]byte(test.input))
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, peers)
	}
}

func TestPeerString(t *testing.T) {
	tests := []struct {
		input  Peer
		output string
	}{
		{
			input:  Peer{IP: net.IP{127, 0, 0, 1}, Port: 1337},
			output: "127.0.0.1:1337",
		},
	}
	for _, test := range tests {
		s := test.input.String()
		assert.Equal(t, test.output, s)
	}
}
