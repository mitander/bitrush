package torrentfile

import (
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
