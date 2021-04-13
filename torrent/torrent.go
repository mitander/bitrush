package main

import (
	"fmt"

	"github.com/mitander/bitrush/peers"
)

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

func (t *Torrent) Download() ([]byte, error) {
	hashes := t.PieceHashes
	queue := make(chan *pieceWork, len(hashes))
	results := make(chan *pieceResult)

	for index, hash := range hashes {
		begin, end := t.pieceBounds(index)
		length := end - begin
		queue <- &pieceWork{index, hash, length}
	}
	for peer := range t.Peers {
		fmt.Println("start worker for: ", peer)
	}

	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(hashes) {
		res := <-results
		begin, end := t.pieceBounds(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++
	}
	close(queue)
	return buf, nil
}

func (t *Torrent) pieceBounds(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}
