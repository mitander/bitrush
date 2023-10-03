package torrent

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/mitander/bitrush/metainfo"
	"github.com/mitander/bitrush/p2p"
	"github.com/mitander/bitrush/storage"
	"github.com/mitander/bitrush/tracker"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

type PeerID [20]byte

func NewPeerID() (PeerID, error) {
	const prefix = "-BR0001-"
	var id PeerID
	copy(id[:8], prefix)
	_, err := rand.Read(id[8:])
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to randomize peer id")
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
	Trackers        []tracker.Tracker
	Peers           []p2p.Peer
	PeerID          [20]byte
	InfoHash        [20]byte
	PieceHashes     [][20]byte
	PieceLength     int
	Length          int
	Name            string
	Files           []storage.File
	LastPeerRequest time.Time
}

func NewTorrent(m *metainfo.MetaInfo) (*Torrent, error) {
	id, err := NewPeerID()
	if err != nil {
		return nil, err
	}

	var trackers []tracker.Tracker
	var peers []p2p.Peer
	for i := range m.Announce {
		tr, err := tracker.NewTracker(m.Announce[i], m.Length, m.InfoHash, id)
		if err != nil {
			return nil, err
		}
		trackers = append(trackers, tr)
	}

	t := &Torrent{
		Trackers:    trackers,
		Peers:       peers,
		PeerID:      id,
		InfoHash:    m.InfoHash,
		PieceHashes: m.PieceHashes,
		PieceLength: m.PieceLength,
		Length:      m.Length,
		Name:        m.Name,
		Files:       m.Files,
	}

	go t.RequestPeers()

	return t, nil
}

func (t *Torrent) Download(path string) error {
	hashes := t.PieceHashes
	queue := make(chan *pieceWork, len(hashes))
	results := make(chan *pieceResult)

	var bar *progressbar.ProgressBar
	render := log.GetLevel() != log.DebugLevel

	if render {
		bar = progressbar.Default(100, "Downloading with 0 workers")
	}

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
	go func() {
		for {
			time.Sleep(2 * time.Second) // TODO: fix this
			for _, peer := range t.Peers {
				if !peer.Active {
					go t.startWorker(peer, queue, results)
				}
			}
		}
	}()

	donePieces := 0
	for donePieces < len(hashes) {
		res := <-results
		begin, _ := t.pieceBounds(res.index)

		sw.Queue <- storage.StorageWork{Data: res.buf, Index: begin}
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

	peer.Active = true
	failures := 0

	for pw := range queue {
		if failures > 3 {
			peer.Active = false
			log.Debugf("peer reached max failures, disconnecting client")
			return
		}
		if !c.Bitfield.HasPiece(pw.index) {
			queue <- pw
			log.Debugf("putting piece '%d' back in queue: not in bitfield", pw.index)
			failures += 1
			continue
		}

		buf, err := c.DownloadPiece(pw.index, pw.length)
		if err != nil {
			queue <- pw
			log.WithFields(log.Fields{"reason": err.Error(), "index": pw.index}).Debug("putting piece back in queue")
			failures += 1
			continue
		}

		err = pw.validate(buf)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "index": pw.index}).Debug("putting piece back in queue")
			queue <- pw
			failures += 1
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

func (t *Torrent) RequestPeers() {
	for {
		for _, tr := range t.Trackers {
			peers, err := tr.RequestPeers()
			if err != nil {
				continue
			}
			t.AppendUnique(peers)
		}
		time.Sleep(20 * time.Second)
		t.LastPeerRequest = time.Now()
	}
}

func (t *Torrent) AppendUnique(p []p2p.Peer) {
	var peers []p2p.Peer
	for _, np := range p {
		exist := false
		for _, kp := range t.Peers {
			if kp.String() == np.String() {
				exist = true
				continue
			}
		}
		if !exist {
			peers = append(peers, np)
		}
	}
	t.Peers = append(t.Peers, peers...)
	log.Debugf("added %d peers, total peers: %d", len(peers), len(t.Peers))
}
