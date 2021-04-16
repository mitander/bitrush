package main

// TODO: implement logger

import (
	"flag"
	"log"
	"os"

	"github.com/mitander/bitrush/logger"
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
		logger.Level(logger.DebugLevel)
	}

	if *read != "" {
		tf, err := torrentfile.OpenFile(*read)
		if err != nil {
			logger.Fatal(err)
		}

		err = tf.Download(*write)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		logger.CLI("BitRush")
		logger.CLI("-------")
		logger.CLI("No .torrent file selected!")
		logger.CLI("Usage: bitrush -f <torrent file>")
		logger.CLI("Help: bitrush -h")
	}

	if *help {
		logger.CLI("")
		logger.CLI("BitRush")
		logger.CLI("-------")
		logger.CLI("-f [file] (required)")
		logger.CLI("Info: torrent file you want to open")
		logger.CLI("Usage: bitrush -f <torrent file>")
		logger.CLI("")
		logger.CLI("-o [out file] (optional)")
		logger.CLI("Info: output file location - default '.' (current directory)")
		logger.CLI("Usage: bitrush -o <output file>")
		logger.CLI("")
		logger.CLI("-h [help] (optional)")
		logger.CLI("Info: show help menu")
		logger.CLI("Usage: bitrush -h")
		logger.CLI("")
		logger.CLI("-d [debug] (optional")
		logger.CLI("info: enable debug")
		logger.CLI("Usage: bitrush -d")
		logger.CLI("")
	}
	logger.CLI("Download finished!")
	logger.CLI("Exiting..")
	os.Exit(1)
}
