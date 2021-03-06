package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	bencode "github.com/jackpal/bencode-go"
	"github.com/mitander/bitrush/logger"
	"github.com/mitander/bitrush/p2p"
	"github.com/mitander/bitrush/peers"
)

const Port uint16 = 6889

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type bencodeInfo struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
}

func (tf *TorrentFile) Download(path string) error {
	peerID, err := peers.GeneratePeerID()
	if err != nil {
		logger.Warning("Error generating peer id")
		return err
	}
	logger.Info("generating peerID")

	peers, err := tf.ReqPeers(peerID, Port)
	if err != nil {
		logger.Warning("Error requesting peers")
		return err
	}
	logger.Info("requesting peers")

	t := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
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

func OpenFile(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		logger.Warning("Error opening file")
		return TorrentFile{}, err

	}
	defer file.Close()
	logger.Info("opening file")

	bct := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bct)
	if err != nil {
		return TorrentFile{}, err
	}
	return bct.toTorrentFile()
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

func (bct *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, pieceHashes, err := bct.Info.hashInfo()
	if err != nil {
		return TorrentFile{}, err
	}
	logger.Info("creating TorrentFile")
	return TorrentFile{
		Announce:    bct.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bct.Info.PieceLength,
		Length:      bct.Info.Length,
		Name:        bct.Info.Name,
	}, nil
}

func (bci *bencodeInfo) hashInfo() ([20]byte, [][20]byte, error) {
	pieces := []byte(bci.Pieces)
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
	err := bencode.Marshal(&info, *bci)
	if err != nil {
		return [20]byte{}, [][20]byte{}, err
	}
	infoHash := sha1.Sum(info.Bytes())
	return infoHash, pieceHashes, nil
}
