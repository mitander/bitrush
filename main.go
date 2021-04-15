package main

// TODO: implement logger

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mitander/bitrush/torrentfile"
)

var (
	read  = flag.String("f", "", "open .torrent file")
	write = flag.String("o", ".", "download location")
	help  = flag.Bool("h", false, "show help")
	debug = flag.Bool("d", false, "enable debug mode")
)

func main() {
	flag.Parse()

	if *debug {
		fmt.Println("Debug enabled")
		// TODO: implement logger
	}

	if *read != "" {
		tf, err := torrentfile.OpenFile(*read)
		if err != nil {
			log.Fatal(err)
		}

		err = tf.Download(*write)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("")
		fmt.Println("BitRush")
		fmt.Println("-------")
		fmt.Println("No .torrent file selected!")
		fmt.Println("")
		fmt.Println("Usage: bitrush -f <torrent file>")
		fmt.Println("Help: bitrush -h")
		fmt.Println("")
	}

	if *help {
		fmt.Println("")
		fmt.Println("BitRush")
		fmt.Println("-------")
		fmt.Println("-f [file] (required)")
		fmt.Println("Info: torrent file you want to open")
		fmt.Println("Usage: bitrush -f <torrent file>")
		fmt.Println("")
		fmt.Println("-o [out file] (optional)")
		fmt.Println("Info: output file location - default '.' (current directory)")
		fmt.Println("Usage: bitrush -o <output file>")
		fmt.Println("")
		fmt.Println("-h [help] (optional)")
		fmt.Println("Info: show help menu")
		fmt.Println("Usage: bitrush -h")
		fmt.Println("")
		fmt.Println("-d [debug] (optional")
		fmt.Println("info: enable debug")
		fmt.Println("Usage: bitrush -d")
		fmt.Println("")
		os.Exit(0)
	}
}
