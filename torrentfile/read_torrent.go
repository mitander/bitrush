package torrentfile

import (
	"fmt"
	"io/ioutil"

	bencode "github.com/IncSW/go-bencode"
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
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return bencodeTorrent{}, err
	}

	tor := bencodeTorrent{}

	val, err := bencode.Unmarshal(file)
	if err != nil {
		return bencodeTorrent{}, err
	}
	fmt.Println(val)
	return tor, nil
}
