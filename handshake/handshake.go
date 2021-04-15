package handshake

import (
	"fmt"
	"io"

	"github.com/mitander/bitrush/peers"
)

//https://wiki.theory.org/BitTorrentSpecification#Handshake
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func New(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// handshake: <pstrlen><pstr><reserved><info_hash><peer_id>
func (h *Handshake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8))
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

func Read(r io.Reader) (*Handshake, error) {
	peerID, err := peers.GeneratePeerID()
	if err != nil {
		return nil, err
	}

	lengthBuf := make([]byte, 1)

	_, err = io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, fmt.Errorf("Error reading buffer")
	}

	pstrlen := int(lengthBuf[0])
	if pstrlen == 0 {
		err := fmt.Errorf("Invalid pstrlen, cannot be 0")
		return nil, err
	}

	hsBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, hsBuf)
	if err != nil {
		return nil, err
	}

	var infoHash [20]byte

	copy(infoHash[:], hsBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], hsBuf[pstrlen+8+20:])

	hs := Handshake{
		Pstr:     string(hsBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}
	return &hs, nil
}
