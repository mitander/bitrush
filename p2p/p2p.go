package p2p

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"runtime"
	"time"

	"github.com/mitander/bitrush/client"
	"github.com/mitander/bitrush/message"
	"github.com/mitander/bitrush/peers"
	log "github.com/sirupsen/logrus"
)

const MaxBlockSize = 16384
const MaxBacklog = 5

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceState struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (t *Torrent) Download() []byte {
	hashes := t.PieceHashes
	queue := make(chan *pieceWork, len(hashes))
	results := make(chan *pieceResult)

	for index, hash := range hashes {
		begin, end := t.pieceBounds(index)
		length := end - begin
		queue <- &pieceWork{index, hash, length}
	}

	log.Info("Download started")
	// for every peer - start a new dowload worker
	for _, peer := range t.Peers {
		go t.startWorker(peer, queue, results)
	}

	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(hashes) {
		res := <-results
		begin, end := t.pieceBounds(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++
		workers := runtime.NumGoroutine() - 1

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		log.Infof("Downloaded: %0.2f%% - Peers: %d\n", percent, workers)
	}
	close(queue)
	log.Debug("closing client")
	return buf
}

func (t *Torrent) startWorker(peer peers.Peer, queue chan *pieceWork, results chan *pieceResult) {
	// create a new client for every peer
	c, err := client.New(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "peer": peer.String()}).Warn("worker failed to create client")
		return
	}
	defer c.Conn.Close()

	c.SendUnchoke()
	c.SendInterested()

	for pw := range queue {
		if !c.Bitfield.HasPiece(pw.index) {
			queue <- pw // put piece back in queue
			log.Debugf("putting piece '%d' back in queue: not in bitfield", pw.index)
			continue
		}

		buf, err := downloadPiece(c, pw)
		if err != nil {
			queue <- pw
			log.WithFields(log.Fields{"reason": err.Error(), "index": pw.index}).Debug("putting piece back in queue")
			continue
		}

		err = validate(pw, buf)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "index": pw.index}).Debug("putting piece back in queue")
			queue <- pw
			continue
		}

		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}

func downloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	// 30 seconds deadline to download piece (262kb)
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	state := pieceState{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
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

func (t *Torrent) pieceBounds(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (state *pieceState) readMessage() error {
	msg, err := message.Read(state.client.Conn)
	if err != nil {
		return err
	}

	if msg == nil { // keep nil messages alive
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		i, err := message.ParseHaveMsg(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(i)
	case message.MsgPiece:
		n, err := message.ParsePieceMsg(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func validate(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		err := errors.New("piece work validation failed")
		log.Error(err.Error())
		return err
	}
	return nil
}
