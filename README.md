# BitRush
A minimal BitTorrent CLI client written in Golang


## Installation

* Install package
```bash
go install "github.com/mitander/bitrush"
```
* Verify GOPATH
```bash
echo $GOPATH
```

* if GOPATH is empty, set GOPATH to $HOME/go (~/go)
```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

## Usage

**CLI**
* Download torrent file
```go
bitrush -f debian.iso.torrent -o debian.iso
```

* Commands
```go
-h help
-f input file
-o output file (default: current directory)
-d debug mode
```

**Library**

* Install package
```go
go get "github.com/mitander/bitrush"
```
* Import package
```go
import "github.com/mitander/bitrush"
```
* Open torrent file
```go
path := "./torrentfile/testdata/debian.torrent"
tf, err := torrentfile.toTorrentFile(path)
if err != nil {
    log.Fatal(err)
}
```

* Download Torrent
```go
path := "debian.iso"
err := tf.Download(path)
if err != nil {
    log.Fatal(err)
}
```

## License
[MIT License](LICENSE).

