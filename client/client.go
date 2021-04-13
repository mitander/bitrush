package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/mitander/bitrush/handshake"
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

	bf, err := bitfield.recvBitfield(conn)
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
