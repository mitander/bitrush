package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mitander/bitrush/metainfo"
	"github.com/mitander/bitrush/torrent"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	read  = flag.String("f", "./metainfo/testdata/debian-10.9.0-amd64-netinst.iso.torrent", "open .torrent file")
	write = flag.String("o", "out", "download directory")
	help  = flag.Bool("h", false, "show help")
	debug = flag.Bool("d", false, "enable debug mode")
)

func main() {
	flag.Parse()

	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *help {
		printHelpMenu()
		os.Exit(1)
	}

	if !strings.Contains(*read, ".torrent") {
		printNoArgs()
		os.Exit(1)
	}

	m, err := metainfo.NewMetaInfo(*read)
	if err != nil {
		log.Fatal(err)
	}

	t, err := torrent.NewTorrent(m)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Download(*write)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Download finished!")
	log.Info("Exiting..")
	os.Exit(1)
}

func printHelpMenu() {
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
	fmt.Println("-------")
	fmt.Println("")
}

func printNoArgs() {
	fmt.Println("")
	fmt.Println("BitRush")
	fmt.Println("-------")
	fmt.Println("No .torrent file selected!")
	fmt.Println("")
	fmt.Println("Usage: bitrush -f <torrent file>")
	fmt.Println("Help: bitrush -h")
	fmt.Println("-------")
	fmt.Println("")
}
