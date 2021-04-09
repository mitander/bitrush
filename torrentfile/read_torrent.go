package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	bencode "github.com/jackpal/bencode-go"
)

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

func OpenFile(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bct := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bct)
	if err != nil {
		return TorrentFile{}, err
	}

	return bct.toTorrentFile()
}

func (bct *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, pieceHashes, err := bct.Info.hashInfo()
	if err != nil {
		fmt.Println(err)
	}

	tf := TorrentFile{
		Announce:    bct.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bct.Info.PieceLength,
		Length:      bct.Info.Length,
		Name:        bct.Info.Name,
	}
	return tf, nil

}

func (bci *bencodeInfo) hashInfo() ([20]byte, [][20]byte, error) {
	pieces := []byte(bci.Pieces)
	hashLen := 20
	numHashes := len(pieces) / hashLen
	pieceHashes := make([][20]byte, numHashes)

	if len(pieces)%hashLen != 0 {
		err := fmt.Errorf("err: fauly pieces length")
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
