package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

		err = tf.Download(".")
		if err != nil {
			log.Fatal(err)
		}

	}

	fmt.Println("No torrent file selected, use -f to select")
	fmt.Println("Exiting")
	os.Exit(0)
}
