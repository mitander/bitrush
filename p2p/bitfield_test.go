package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasPiece(t *testing.T) {
	bf := Bitfield{0b01000110, 0b01010100}
	outputs := []bool{false, true, false, false, false, true, true, false, false, true, false, true, false, true, false, false, false, false, false, false}
	for i := 0; i < len(outputs); i++ {
		assert.Equal(t, outputs[i], bf.HasPiece(i))
	}
}

func TestSetPiece(t *testing.T) {
	tests := []struct {
		input  Bitfield
		index  int
		output Bitfield
	}{
		{
			input:  Bitfield{0b01010110, 0b01010100},
			index:  4,
			output: Bitfield{0b01011110, 0b01010100},
		},
		{
			input:  Bitfield{0b01010110, 0b01010100},
			index:  9,
			output: Bitfield{0b01010110, 0b01010100},
		},
		{
			input:  Bitfield{0b01010100, 0b01010100},
			index:  15,
			output: Bitfield{0b01010100, 0b01010101},
		},
		{
			input:  Bitfield{0b01010110, 0b01110100},
			index:  18,
			output: Bitfield{0b01010110, 0b01110100},
		},
	}
	for _, test := range tests {
		bf := test.input
		bf.SetPiece(test.index)
		assert.Equal(t, test.output, bf)
	}
}
