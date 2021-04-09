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

	fmt.Println(tf)
}
