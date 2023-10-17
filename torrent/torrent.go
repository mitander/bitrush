package torrent

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"time"

	"github.com/mitander/bitrush/metainfo"
	"github.com/mitander/bitrush/peer"
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
	Peers           []peer.Peer
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
	downloadC       chan *pieceWork
	resultC         chan *pieceResult
	workerC         chan peer.Peer
	ActiveWorkers   uint
}

func NewTorrent(m *metainfo.MetaInfo) (*Torrent, error) {
	id, err := NewPeerID()
	if err != nil {
		return nil, err
	}

	var trackers []tracker.Tracker
	var peers []peer.Peer
	for i := range m.Announce {
		tr, err := tracker.NewTracker(m.Announce[i], m.Length, m.InfoHash, id)
		if err != nil {
			return nil, err
		}
		trackers = append(trackers, tr)
	}

	t := &Torrent{
		Trackers:      trackers,
		Peers:         peers,
		PeerID:        id,
		InfoHash:      m.InfoHash,
		PieceHashes:   m.PieceHashes,
		PieceLength:   m.PieceLength,
		Length:        m.Length,
		Name:          m.Name,
		Files:         m.Files,
		downloadC:     make(chan *pieceWork, len(m.PieceHashes)),
		resultC:       make(chan *pieceResult),
		workerC:       make(chan peer.Peer),
		ActiveWorkers: 0,
	}

	return t, nil
}

func (t *Torrent) Download(path string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sw, err := storage.NewStorageWorker(ctx, path, t.Files)
	if err != nil {
		return err
	}
	go sw.StartWorker()

	go func() {
		for index, hash := range t.PieceHashes {
			begin, end := t.pieceBounds(index)
			length := end - begin
			t.downloadC <- &pieceWork{index, hash, length}
		}
	}()

	log.Info("Download started")

	go t.RequestPeers(ctx)
	go t.PeerDownload(ctx)

	for t.Downloaded < len(t.PieceHashes) {
		res := <-t.resultC
		begin, _ := t.pieceBounds(res.index)

		sw.Queue <- storage.StorageWork{Data: res.buf, Index: begin}
		t.Downloaded++

		t.Progress = float64(t.Downloaded) / float64(len(t.PieceHashes)) * 100
		log.Debugf("Downloaded: %0.2f%% - Peers: %d", t.Progress, t.ActiveWorkers)
	}

	err = sw.Complete()
	if err != nil {
		log.Errorf("failed to complete storage work: %s", err.Error())
		return err
	}

	return nil
}

func (t *Torrent) PeerDownload(ctx context.Context) {
	for {
		select {
		case p := <-t.workerC:
			go t.startWorker(ctx, p)
		case <-ctx.Done():
			return
		}
	}
}

func (t *Torrent) startWorker(ctx context.Context, p peer.Peer) {
	cooldown := 5 * time.Second
	c, err := peer.NewClient(p, t.PeerID, t.InfoHash)
	if err != nil {
		time.Sleep(cooldown)
		t.workerC <- p
		return
	}
	defer c.Conn.Close()
	t.ActiveWorkers++

	c.SendUnchoke()
	c.SendInterested()

	for {
		select {
		case pw := <-t.downloadC:
			if !c.Bitfield.HasPiece(pw.index) {
				t.downloadC <- pw
				log.Debugf("putting piece '%d' back in queue: not in bitfield", pw.index)
				time.Sleep(cooldown)
				continue
			}

			buf, err := c.DownloadPiece(pw.index, pw.length)
			if err != nil {
				t.downloadC <- pw
				log.WithFields(log.Fields{"reason": err.Error(), "index": pw.index}).Debug("putting piece back in queue")
				time.Sleep(cooldown)
				continue
			}

			err = pw.validate(buf)
			if err != nil {
				t.downloadC <- pw
				log.WithFields(log.Fields{"reason": err.Error(), "index": pw.index}).Debug("putting piece back in queue")
				time.Sleep(cooldown)
				continue
			}

			c.SendHave(pw.index)
			t.resultC <- &pieceResult{pw.index, buf}

		case <-ctx.Done():
			t.ActiveWorkers--
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
		// request new peers from trackers every 5 seconds
		case <-time.After(time.Until(t.LastPeerRequest.Add(5 * time.Second))):
			for _, tr := range t.Trackers {
				peers, err := tr.RequestPeers()
				if err != nil {
					continue
				}
				peers = t.FilterUnique(peers)
				t.Peers = append(t.Peers, peers...)
				log.Debugf("added %d peers, total peers: %d", len(peers), len(t.Peers))

				for _, p := range peers {
					t.workerC <- p
				}
			}
			t.LastPeerRequest = time.Now()

		case <-ctx.Done():
			return
		}
	}
}

func (t *Torrent) FilterUnique(p []peer.Peer) []peer.Peer {
	var peers []peer.Peer
	for _, np := range p {
		exist := false
		for _, kp := range t.Peers {
			if kp.String() == np.String() {
				exist = true
				break
			}
		}
		if !exist {
			peers = append(peers, np)
		}
	}
	return peers
}
