package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mitander/bitrush/torrentfile"
)

func main() {
	inPath := os.Args[1]

	tf, err := torrentfile.OpenFile(inPath)
	if err != nil {
		log.Fatal(err)
	}

	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	peers, err := tf.ReqPeers(peerID, torrentfile.Port)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(peers)
}
