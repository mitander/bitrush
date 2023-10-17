package peer

import (
	"net"
	"time"

	"github.com/mitander/bitrush/bitfield"
	"github.com/mitander/bitrush/handshake"
	"github.com/mitander/bitrush/message"
	log "github.com/sirupsen/logrus"
)

const MaxBlockSize = 16384
const MaxBacklog = 5

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     Peer
	infoHash [20]byte
	peerID   [20]byte
}

func NewClient(peer Peer, peerID, infoHash [20]byte) (*Client, error) {
	// should not take more than 3 seconds to establish connection
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	hs := handshake.NewHandshake(infoHash, peerID)
	_, err = hs.Send(conn)
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
	return c.send(message.FormatRequestMsg(index, begin, length))
}

func (c *Client) SendHave(index int) error {
	return c.send(message.FormatHaveMsg(index))
}

func (c *Client) SendInterested() error {
	return c.send(&message.Message{ID: message.MsgInterested})
}

func (c *Client) SendNotInterested() error {
	return c.send(&message.Message{ID: message.MsgNotInterested})
}

func (c *Client) SendUnchoke() error {
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

func (c *Client) DownloadPiece(index int, length int) ([]byte, error) {
	// 30 seconds deadline to download piece (262kb)
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	state := pieceState{
		index:  index,
		client: c,
		buf:    make([]byte, length),
	}

	for state.downloaded < length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < length {
				blockSize := MaxBlockSize
				if length-state.requested < blockSize {
					blockSize = length - state.requested
				}

				err := c.SendRequest(index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}
		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

type pieceState struct {
	index      int
	client     *Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (s *pieceState) readMessage() error {
	msg, err := message.ReadMessage(s.client.Conn)
	if err != nil {
		return err
	}

	switch msg.ID {
	case message.MsgKeepAlive:
		break
	case message.MsgUnchoke:
		s.client.Choked = false
	case message.MsgChoke:
		s.client.Choked = true
	case message.MsgHave:
		i, err := message.ParseHaveMsg(msg)
		if err != nil {
			return err
		}
		s.client.Bitfield.SetPiece(i)
	case message.MsgPiece:
		n, err := message.ParsePieceMsg(s.index, s.buf, msg)
		if err != nil {
			return err
		}
		s.downloaded += n
		s.backlog--
	}
	return nil
}
