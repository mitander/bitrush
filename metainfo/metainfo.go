package metainfo

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"os"
	"path/filepath"

	bencode "github.com/jackpal/bencode-go"
	"github.com/mitander/bitrush/p2p"
	"github.com/mitander/bitrush/peers"
	"github.com/mitander/bitrush/tracker"
	log "github.com/sirupsen/logrus"
)

type MetaInfo struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type torrentBencode struct {
	Announce string      `bencode:"announce"`
	Info     infoBencode `bencode:"info"`
}

type infoBencode struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
}

func (m *MetaInfo) Download(path string) error {
	peerID, err := peers.GeneratePeerID()
	if err != nil {
		return err
	}
	log.Debug("generating peer id")

	tr, err := tracker.NewTracker(m.Announce, m.Length, m.InfoHash, peerID)
	if err != nil {
		return err
	}

	log.Debug("requesting peers")
	peers, err := tr.ReqPeers()
	if err != nil {
		return err
	}

	t := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    m.InfoHash,
		PieceHashes: m.PieceHashes,
		PieceLength: m.PieceLength,
		Length:      m.Length,
		Name:        m.Name,
	}

	path = filepath.Join(path, m.Name)
	err = t.Download(path)
	if err != nil {
		return err
	}

	log.Infof("Writing torrent to file: %s", path)
	return nil
}

func FromFile(path string) (*MetaInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("failed to open file")
		return nil, err

	}
	defer file.Close()

	bct := torrentBencode{}
	err = bencode.Unmarshal(file, &bct)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("failed to unmarshal bencode from file")
		return nil, err
	}
	return bct.toMetaInfo()
}

func (bct *torrentBencode) toMetaInfo() (*MetaInfo, error) {
	infoHash, pieceHashes, err := bct.Info.hash()
	if err != nil {
		return nil, err
	}
	m := &MetaInfo{
		Announce:    bct.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bct.Info.PieceLength,
		Length:      bct.Info.Length,
		Name:        bct.Info.Name,
	}
	log.Debugf("created torrent meta info: %s", bct.Info.Name)
	return m, nil
}

func (i *infoBencode) hash() ([20]byte, [][20]byte, error) {
	pieces := []byte(i.Pieces)
	hashLen := 20
	numHashes := len(pieces) / hashLen
	pieceHashes := make([][20]byte, numHashes)

	if len(pieces)%hashLen != 0 {
		err := errors.New("invalid hash length")
		log.Error(err.Error())
		return [20]byte{}, [][20]byte{}, err
	}
	for i := range pieceHashes {
		copy(pieceHashes[i][:], pieces[i*hashLen:(i+1)*hashLen])
	}

	var info bytes.Buffer
	err := bencode.Marshal(&info, *i)
	if err != nil {
		return [20]byte{}, [][20]byte{}, err
	}
	infoHash := sha1.Sum(info.Bytes())
	return infoHash, pieceHashes, nil
}
