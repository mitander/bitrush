package client

import (
	"net"
	"time"

	"github.com/mitander/bitrush/bitfield"
	"github.com/mitander/bitrush/handshake"
	"github.com/mitander/bitrush/message"
	"github.com/mitander/bitrush/peers"
	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{"reason": err.Error(), "peer": peer.String()}).Error("failed to connect to peer")
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

func (c *Client) SendRequest(index, begin, length int) error {
	log.Debug("sending 'request'")
	return c.send(message.FormatRequestMsg(index, begin, length))
}

func (c *Client) SendHave(index int) error {
	log.Debug("sending 'have'")
	return c.send(message.FormatHaveMsg(index))
}

func (c *Client) SendInterested() error {
	log.Debug("sending 'interested'")
	return c.send(&message.Message{ID: message.MsgInterested})
}

func (c *Client) SendNotInterested() error {
	log.Debug("sending 'not interested'")
	return c.send(&message.Message{ID: message.MsgNotInterested})
}

func (c *Client) SendUnchoke() error {
	log.Debug("sending 'unchoke'")
	return c.send(&(message.Message{ID: message.MsgUnchoke}))
}

func (c *Client) send(msg *message.Message) error {
	_, err := c.Conn.Write(msg.Serialize())
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "message": msg.String()}).Error("failed to send message")
		return err
	}
	return nil
}
