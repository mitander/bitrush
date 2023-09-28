package torrent

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"runtime"

	"github.com/mitander/bitrush/metainfo"
	"github.com/mitander/bitrush/p2p"
	"github.com/mitander/bitrush/storage"
	"github.com/mitander/bitrush/tracker"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// http://www.bittorrent.org/beps/bep_0020.html
var peerIDPrefix = []byte("-BR0001-")

type PeerID [20]byte

func NewPeerID() (PeerID, error) {
	var id PeerID
	copy(id[:], peerIDPrefix)
	_, err := rand.Read(id[len(peerIDPrefix):])
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to generate peer id")
		return PeerID{}, err
	}
	return id, nil
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

func (p *pieceWork) validate(buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], p.hash[:]) {
		err := errors.New("piece work validation failed")
		log.Error(err.Error())
		return err
	}
	return nil
}

type Torrent struct {
	Trackers    []tracker.Tracker
	Peers       []p2p.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	Files       []storage.File
}

func NewTorrent(m *metainfo.MetaInfo) (*Torrent, error) {
	peerID, err := NewPeerID()
	if err != nil {
		return nil, err
	}

	var peers []p2p.Peer
	for i := range m.Announce {
		tr, err := tracker.NewTracker(m.Announce[i], m.Length, m.InfoHash, peerID)
		if err != nil {
			return nil, err
		}
		p, err := tr.ReqPeers()
		peers = append(peers, p...)
		if err != nil {
			return nil, err
		}
	}

	return &Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    m.InfoHash,
		PieceHashes: m.PieceHashes,
		PieceLength: m.PieceLength,
		Length:      m.Length,
		Name:        m.Name,
		Files:       m.Files,
	}, nil
}

func (t *Torrent) Download(path string) error {
	hashes := t.PieceHashes
	queue := make(chan *pieceWork, len(hashes))
	results := make(chan *pieceResult)

	sw, err := storage.NewStorageWorker(path, t.Files)
	if err != nil {
		return err
	}
	go sw.StartWorker()

	for index, hash := range hashes {
		begin, end := t.pieceBounds(index)
		length := end - begin
		queue <- &pieceWork{index, hash, length}
	}

	log.Info("Download started")
	for _, peer := range t.Peers {
		go t.startWorker(peer, queue, results)
	}

	var bar *progressbar.ProgressBar
	render := log.GetLevel() != log.DebugLevel

	if render {
		bar = progressbar.Default(100, "Downloading with 0 workers")
	}

	donePieces := 0
	for donePieces < len(hashes) {
		res := <-results
		begin, _ := t.pieceBounds(res.index)

		sw.Queue <- storage.StorageWork{Data: res.buf, Index: int64(begin)}
		donePieces++

		if render {
			bar.Describe(fmt.Sprintf("Downloading with %d workers", runtime.NumGoroutine()-2))
			bar.Set(int(float64(donePieces) / float64(len(t.PieceHashes)) * 100))
		}
	}
	sw.Exit <- 0
	close(queue)
	close(results)
	return nil
}

func (t *Torrent) startWorker(peer p2p.Peer, queue chan *pieceWork, results chan *pieceResult) {
	c, err := p2p.NewClient(peer, t.PeerID, t.InfoHash)
	if err != nil {
		return
	}
	defer c.Conn.Close()

	c.SendUnchoke()
	c.SendInterested()

	for pw := range queue {
		if !c.Bitfield.HasPiece(pw.index) {
			queue <- pw
			log.Debugf("putting piece '%d' back in queue: not in bitfield", pw.index)
			continue
		}

		buf, err := c.DownloadPiece(pw.index, pw.length)
		if err != nil {
			queue <- pw
			log.WithFields(log.Fields{"reason": err.Error(), "index": pw.index}).Debug("putting piece back in queue")
			continue
		}

		err = pw.validate(buf)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "index": pw.index}).Debug("putting piece back in queue")
			queue <- pw
			continue
		}

		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}

func (t *Torrent) pieceBounds(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}
