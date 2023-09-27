package metainfo

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"os"

	bencode "github.com/jackpal/bencode-go"
	log "github.com/sirupsen/logrus"
)

type MetaInfo struct {
	Announce    []string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	AnnounceList []string    `bencode:"announce-list"`
	Info         bencodeInfo `bencode:"info"`
}

type bencodeInfo struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
}

func NewMetaInfo(path string) (*MetaInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("failed to open file")
		return nil, err

	}
	defer file.Close()

	bct := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bct)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("failed to unmarshal bencode from file")
		return nil, err
	}
	return bct.toMetaInfo()
}

func (bct *bencodeTorrent) toMetaInfo() (*MetaInfo, error) {
	infoHash, pieceHashes, err := bct.Info.hash()
	if err != nil {
		return nil, err
	}

	var announce []string
	if len(bct.AnnounceList) != 0 {
		announce = append(announce, bct.AnnounceList...)
	} else {
		announce = append(bct.AnnounceList, bct.Announce)
	}

	m := &MetaInfo{
		Announce:    announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bct.Info.PieceLength,
		Length:      bct.Info.Length,
		Name:        bct.Info.Name,
	}
	log.Debugf("created torrent meta info: %s", bct.Info.Name)

	return m, nil
}

func (i *bencodeInfo) hash() ([20]byte, [][20]byte, error) {
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
