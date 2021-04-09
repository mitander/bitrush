package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mitander/bitrush/torrentfile"
)

func main() {
	inPath := os.Args[1]

	tor, err := torrentfile.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tor)
}
