# BitRush
A minimal BitTorrent CLI client written in Golang


## Installation

`$ go get "github.com/mitander/bitrush"`

## How to use


**Import package**
```go
 import "github.com/mitander/bitrush"
```


**Open torrent file**
```go
path := "./torrentfile/testdata/debian.torrent"
tf, err := torrentfile.toTorrentFile(path)
if err != nil {
    log.Fatal(err)
}
```

**Download**
```go
path := "debian.iso"
err := tf.Download(path)
if err != nil {
    log.Fatal(err)
}
```

## License
[MIT License](LICENSE).
