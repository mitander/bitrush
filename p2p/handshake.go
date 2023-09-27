package p2p

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// https://wiki.theory.org/BitTorrentSpecification#Handshake
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash [20]byte, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Send(conn net.Conn) (*Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	_, err := conn.Write(h.Serialize())
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to send handshake")
		return nil, err
	}
	res, err := ReadHandshake(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], h.InfoHash[:]) {
		err := errors.New("invalid info hash")
		log.WithFields(log.Fields{"got": res.InfoHash, "expected": h.InfoHash}).Error(err.Error())
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

func ReadHandshake(r io.Reader) (*Handshake, error) {
	l := make([]byte, 1)
	_, err := io.ReadFull(r, l)
	if err != nil {
		return nil, err
	}

	pstrlen := int(l[0])
	if pstrlen == 0 {
		err := errors.New("pstrlen cannot be 0 ")
		return nil, err
	}

	buf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	var infoHash [20]byte
	var peerID [20]byte

	copy(infoHash[:], buf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], buf[pstrlen+8+20:])

	return &Handshake{
		Pstr:     string(buf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}
