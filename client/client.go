package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/mitander/bitrush/bitfield"
	"github.com/mitander/bitrush/handshake"
	"github.com/mitander/bitrush/message"
	"github.com/mitander/bitrush/peers"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func New(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
	// should not take more than 3 seconds to establish connection
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = handshake.DoHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := bitfield.RecvBitfield(conn)
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

func DoHandshake(conn net.Conn, infohash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	hs := handshake.New(infohash, peerID)

	_, err := conn.Write(hs.Serialize())
	if err != nil {
		return nil, fmt.Errorf("handshake failed: writing connection")
	}
	res, err := handshake.Read(conn)
	if err != nil {
		return nil, fmt.Errorf("handshake failed: reading connection")
	}
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("handshake failed: invalid infohash (recieved: %x - expected: %x)", res.InfoHash, infohash)
	}
	return res, nil
}

func (c *Client) SendRequest(index, begin, length int) error {
	req := message.FormatRequestMsg(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendHave(index int) error {
	msg := message.FormatHaveMsg(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
