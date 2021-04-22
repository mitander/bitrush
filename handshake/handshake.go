package handshake

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

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

func DoHandshake(conn net.Conn, infohash, peerID [20]byte) (*Handshake, error) {
	// should not take more than 3 seconds to complete handshake
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	// create new handshake
	hs := New(infohash, peerID)

	// write to connection
	_, err := conn.Write(hs.Serialize())
	if err != nil {
		return nil, fmt.Errorf("handshake failed: writing connection")
	}
	// read from connection
	res, err := Read(conn)
	if err != nil {
		return nil, fmt.Errorf("handshake failed: reading connection")
	}
	// compare info-hashes
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("handshake failed: invalid infohash (recieved: %x - expected: %x)", res.InfoHash, infohash)
	}
	return res, nil
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
		return nil, fmt.Errorf("error reading buffer")
	}

	pstrlen := int(lengthBuf[0])
	if pstrlen == 0 {
		err := fmt.Errorf("invalid pstrlen: cannot be 0")
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
