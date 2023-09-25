package metainfo

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	bencode "github.com/jackpal/bencode-go"
	"github.com/mitander/bitrush/logger"
	"github.com/mitander/bitrush/p2p"
	"github.com/mitander/bitrush/peers"
	"github.com/mitander/bitrush/tracker"
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
		logger.Warning("Error generating peer id")
		return err
	}
	logger.Info("generating peerID")

	tr, err := tracker.NewTracker(m.Announce, m.Length, m.InfoHash, peerID)
	peers, err := tr.ReqPeers()
	if err != nil {
		logger.Warning("Error requesting peers")
		return err
	}
	logger.Info("requesting peers")

	t := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    m.InfoHash,
		PieceHashes: m.PieceHashes,
		PieceLength: m.PieceLength,
		Length:      m.Length,
		Name:        m.Name,
	}

	buf, err := t.Download()
	if err != nil {
		logger.Warning("Error downloading torrent")
		return err
	}

	err = WriteFile(path, buf)
	if err != nil {
		logger.Warning("Error writing to file")
		return err
	}
	logger.CLI(fmt.Sprintf("Writing torrent to file: %s", path))
	return nil
}

func OpenFile(path string) (MetaInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		logger.Warning("Error opening file")
		return MetaInfo{}, err

	}
	defer file.Close()
	logger.Info("opening file")

	bct := torrentBencode{}
	err = bencode.Unmarshal(file, &bct)
	if err != nil {
		return MetaInfo{}, err
	}
	return bct.toMetaInfo()
}

func WriteFile(path string, buf []byte) error {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println(path)
		return err
	}
	defer file.Close()
	_, err = file.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func (bct *torrentBencode) toMetaInfo() (MetaInfo, error) {
	infoHash, pieceHashes, err := bct.Info.hash()
	if err != nil {
		return MetaInfo{}, err
	}
	logger.Info("creating TorrentFile")
	return MetaInfo{
		Announce:    bct.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bct.Info.PieceLength,
		Length:      bct.Info.Length,
		Name:        bct.Info.Name,
	}, nil
}

func (i *infoBencode) hash() ([20]byte, [][20]byte, error) {
	pieces := []byte(i.Pieces)
	hashLen := 20
	numHashes := len(pieces) / hashLen
	pieceHashes := make([][20]byte, numHashes)

	if len(pieces)%hashLen != 0 {
		err := fmt.Errorf("reading hash info failed: invalid hash length (length: %d - expected: %d", len(pieces), hashLen)
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
