package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mitander/bitrush/peers"
	"github.com/mitander/bitrush/torrentfile"
)

var (
	from  = flag.String("f", "", "open .torrent file")
	where = flag.String("w", ".", "download location")
	debug = flag.Bool("d", false, "enable debug mode")
)

func main() {
	flag.Parse()

	if *debug {
		fmt.Println("Debug enabled")
		// TODO: implement logger
	}

	if *from != "" {

		tf, err := torrentfile.OpenFile(*from)
		if err != nil {
			log.Fatal(err)
		}

		peerID, err := peers.GeneratePeerID()
		if err != nil {
			log.Fatal(err)
		}

		peers, err := tf.ReqPeers(peerID, torrentfile.Port)
		if err != nil {
			log.Fatal(err)
		}

		for _, peer := range peers {
			fmt.Println(peer)
		}
	}

	fmt.Println("No torrent file selected, use -f to select")
	fmt.Println("Exiting")
	os.Exit(0)
}
