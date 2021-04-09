package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	bencode "github.com/jackpal/bencode-go"
)

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

func Open(path string) (bencodeTorrent, error) {
	file, err := os.Open(path)
	if err != nil {
		return bencodeTorrent{}, err
	}
	defer file.Close()

	bct := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bct)
	if err != nil {
		return bencodeTorrent{}, err
	}

	return bct, nil
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
