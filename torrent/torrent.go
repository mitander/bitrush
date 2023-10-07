package torrent

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"time"

	"github.com/mitander/bitrush/metainfo"
	"github.com/mitander/bitrush/p2p"
	"github.com/mitander/bitrush/storage"
	"github.com/mitander/bitrush/tracker"
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
		return errors.New("piece work validation failed")
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
	Progress        float64
	Downloaded      int
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

	return t, nil
}

func (t *Torrent) Download(path string) error {
	hashes := t.PieceHashes
	queue := make(chan *pieceWork, len(hashes))
	results := make(chan *pieceResult)
	defer close(queue)
	defer close(results)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sw, err := storage.NewStorageWorker(ctx, path, t.Files)
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

	go t.RequestPeers(ctx)
	go t.PeerDownload(ctx, queue, results)

	for t.Downloaded < len(hashes) {
		res := <-results
		begin, _ := t.pieceBounds(res.index)

		sw.Queue <- storage.StorageWork{Data: res.buf, Index: begin}
		t.Downloaded++

		peers := t.GetActivePeerCount()
		t.Progress = float64(t.Downloaded) / float64(len(t.PieceHashes)) * 100
		log.Debugf("Downloaded: %0.2f%% - Peers: %d", t.Progress, peers)
	}

	err = sw.Complete()
	if err != nil {
		log.Errorf("failed to complete storage work: %s", err.Error())
		return err
	}

	return nil
}

func (t *Torrent) PeerDownload(ctx context.Context, queue chan *pieceWork, results chan *pieceResult) {
	for {
		select {
		case <-time.After(1 * time.Second):
			for i := range t.Peers {
				peer := &t.Peers[i]
				if !peer.Active {
					peer.Active = true
					go t.startWorker(ctx, peer, queue, results)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (t *Torrent) startWorker(ctx context.Context, peer *p2p.Peer, queue chan *pieceWork, results chan *pieceResult) {
	defer func() {
		peer.Active = false
	}()

	c, err := p2p.NewClient(*peer, t.PeerID, t.InfoHash)
	if err != nil {
		return
	}
	defer c.Conn.Close()

	c.SendUnchoke()
	c.SendInterested()

	failures := 0

	for {
		select {
		case pw := <-queue:
			if failures > 3 {
				queue <- pw
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

		case <-ctx.Done():
			return
		}
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

func (t *Torrent) RequestPeers(ctx context.Context) {
	for {
		select {
		// request new peers from trackers every 20 seconds
		case <-time.After(time.Until(t.LastPeerRequest.Add(20 * time.Second))):
			for _, tr := range t.Trackers {
				peers, err := tr.RequestPeers()
				if err != nil {
					continue
				}
				t.AppendUnique(peers)
			}
			t.LastPeerRequest = time.Now()

		case <-ctx.Done():
			return
		}
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

func (t *Torrent) GetActivePeerCount() int {
	peers := t.Peers
	count := 0
	for _, p := range peers {
		if p.Active {
			count += 1
		}
	}
	return count
}
