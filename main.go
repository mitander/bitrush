package main

import (
	"flag"
	"log"
	"os"

	"github.com/mitander/bitrush/logger"
	"github.com/mitander/bitrush/metainfo"
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

	if *help {
		logger.Help()
		os.Exit(1)
	}

	if *read != "" {
		tf, err := metainfo.OpenFile(*read)
		if err != nil {
			logger.Fatal(err)
		}

		err = tf.Download(*write)
		if err != nil {
			log.Fatal(err)
		}

		logger.CLI("Download finished!")
		logger.CLI("Exiting..")
		os.Exit(1)
	} else {
		logger.NoArgs()
		os.Exit(1)
	}
}
