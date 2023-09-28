package metainfo

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"os"

	bencode "github.com/jackpal/bencode-go"
	"github.com/mitander/bitrush/storage"
	log "github.com/sirupsen/logrus"
)

type MetaInfo struct {
	Announce    []string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	Files       []storage.File
}

type bencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	AnnounceList []string    `bencode:"announce-list"`
	Info         bencodeInfo `bencode:"info"`
}

type bencodeFile struct {
	Path   string `bencode:"path"`
	Length int    `bencode:"length"`
}

type bencodeInfo struct {
	Name        string        `bencode:"name"`
	Length      int           `bencode:"length"`
	Pieces      string        `bencode:"pieces"`
	PieceLength int           `bencode:"piece length"`
	Files       []bencodeFile `bencode:"files,omitempty"`
}

func NewMetaInfo(path string) (*MetaInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("failed to open file")
		return nil, err

	}
	defer file.Close()

	bt := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bt)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("failed to unmarshal bencode from file")
		return nil, err
	}
	return bt.toMetaInfo()
}

func (bt *bencodeTorrent) toMetaInfo() (*MetaInfo, error) {
	infoHash, pieceHashes, err := bt.Info.hash()
	if err != nil {
		return nil, err
	}

	var announce []string
	if len(bt.AnnounceList) != 0 {
		announce = append(announce, bt.AnnounceList...)
	} else {
		announce = append(bt.AnnounceList, bt.Announce)
	}

	var length int
	var files []storage.File
	if len(bt.Info.Files) != 0 {
		files = append(files, storage.File{Path: bt.Info.Name, Length: 0})
		for _, f := range bt.Info.Files {
			files = append(files, storage.File{Path: f.Path, Length: f.Length})
			length += f.Length
		}
		log.Errorf("%v", infoHash)
	} else {
		files = append(files, storage.File{Path: bt.Info.Name, Length: bt.Info.Length})
		length = bt.Info.Length
	}

	m := &MetaInfo{
		Announce:    announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bt.Info.PieceLength,
		Length:      length,
		Name:        bt.Info.Name,
		Files:       files,
	}
	log.Debugf("created torrent meta info: %s", bt.Info.Name)

	return m, nil
}

func (bi *bencodeInfo) hash() ([20]byte, [][20]byte, error) {
	pieces := []byte(bi.Pieces)
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
	err := bencode.Marshal(&info, *bi)
	if err != nil {
		return [20]byte{}, [][20]byte{}, err
	}
	infoHash := sha1.Sum(info.Bytes())
	return infoHash, pieceHashes, nil
}
