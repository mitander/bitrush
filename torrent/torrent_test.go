package torrent

import (
	"net"
	"testing"

	"github.com/mitander/bitrush/peer"
	"github.com/stretchr/testify/assert"
)

func TestAppendUnique(t *testing.T) {
	tests := map[string]struct {
		exist  []peer.Peer
		new    []peer.Peer
		expect []peer.Peer
	}{
		"test 1": {
			exist: []peer.Peer{
				{IP: net.IP{192, 168, 1, 0}, Port: 1337},
				{IP: net.IP{192, 168, 1, 1}, Port: 1337},
			},
			new: []peer.Peer{
				{IP: net.IP{192, 168, 1, 0}, Port: 1337},
				{IP: net.IP{192, 168, 1, 1}, Port: 1337},
			},
			expect: nil,
		},
		"test 2": {
			exist: []peer.Peer{
				{IP: net.IP{192, 168, 1, 0}, Port: 1337},
				{IP: net.IP{192, 168, 1, 1}, Port: 1337},
			},
			new: []peer.Peer{
				{IP: net.IP{192, 168, 1, 1}, Port: 1337},
				{IP: net.IP{192, 168, 1, 2}, Port: 1337},
			},
			expect: []peer.Peer{
				{IP: net.IP{192, 168, 1, 2}, Port: 1337},
			},
		},
		"test 3": {
			exist: []peer.Peer{
				{IP: net.IP{192, 168, 1, 0}, Port: 1337},
				{IP: net.IP{192, 168, 1, 1}, Port: 1337},
			},
			new: []peer.Peer{
				// allow multiple ports on same ip,
				// not sure if this is needed
				{IP: net.IP{192, 168, 1, 1}, Port: 1338},
				{IP: net.IP{192, 168, 1, 2}, Port: 1337},
			},
			expect: []peer.Peer{
				{IP: net.IP{192, 168, 1, 1}, Port: 1338},
				{IP: net.IP{192, 168, 1, 2}, Port: 1337},
			},
		},
	}
	for name, test := range tests {
		torrent := &Torrent{Peers: test.exist}
		got := torrent.filterUnique(test.new)
		assert.Equal(t, test.expect, got, name)
	}
}

func TestPieceBounds(t *testing.T) {
	tests := map[string]struct {
		pieceLength int
		length      int
		index       int
		begin       int
		end         int
	}{
		"test 1": {
			pieceLength: 70,
			length:      400,
			index:       5,
			begin:       350,
			end:         400,
		},
		"test 2": {
			pieceLength: 70,
			length:      420,
			index:       5,
			begin:       350,
			end:         420,
		},
		"test 3": {
			pieceLength: 100,
			length:      450,
			index:       4,
			begin:       400,
			end:         450,
		},
	}
	for name, test := range tests {
		torrent := &Torrent{Length: test.length, PieceLength: test.pieceLength}
		begin, end := torrent.pieceBounds(test.index)
		assert.Equal(t, test.begin, begin, name)
		assert.Equal(t, test.end, end, name)
	}
}
