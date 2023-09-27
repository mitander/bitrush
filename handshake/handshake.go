package handshake

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"

	"github.com/mitander/bitrush/peers"
	log "github.com/sirupsen/logrus"
)

// https://wiki.theory.org/BitTorrentSpecification#Handshake
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
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	hs := New(infohash, peerID)

	_, err := conn.Write(hs.Serialize())
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to send handshake")
		return nil, err
	}
	res, err := read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		err := errors.New("invalid info hash")
		log.WithFields(log.Fields{"got": res.InfoHash, "expected": infohash}).Error(err.Error())
		return nil, err
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

func read(r io.Reader) (*Handshake, error) {
	peerID, err := peers.GeneratePeerID()
	if err != nil {
		return nil, err
	}

	bufLen := make([]byte, 1)

	_, err = io.ReadFull(r, bufLen)
	if err != nil {
		return nil, err
	}

	pstrlen := int(bufLen[0])
	if pstrlen == 0 {
		err := errors.New("pstrlen cannot be 0 ")
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
