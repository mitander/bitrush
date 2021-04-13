package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/mitander/bitrush/handshake"
	"github.com/mitander/bitrush/message"
	"github.com/mitander/bitrush/peers"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield []byte
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

type Bitfield []byte

func New(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
	// set dial timeout to 3 seconds
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	err = doHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}

func doHandshake(conn net.Conn, infohash, peerID [20]byte) error {
	// set deadline to fail instead of blocking after 3 seconds
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	hs := handshake.New(infohash, peerID)
	_, err := conn.Write(hs.Serialize())
	if err != nil {
		return err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return err
	}
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return fmt.Errorf("Invalid infohash recieved: %x - expected: %x", res.InfoHash, infohash)
	}
	return nil
}

func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}
	return bf[byteIndex]>>uint(7-offset)&1 != 0
}

func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	// silently discard invalid bounded index
	if byteIndex < 0 || byteIndex >= len(bf) {
		return
	}
	bf[byteIndex] |= 1 << uint(7-offset)
}

func recvBitfield(conn net.Conn) (Bitfield, error) {
	// set deadline to fail instead of blocking after 3 seconds
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("Invalid bitfield data: %s", msg)
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Invalid msg ID - value: %d expected: 5", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}
