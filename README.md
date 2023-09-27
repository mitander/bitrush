# bitrush
A minimal BitTorrent library written in Go

## Disclaimer
This project is work in progress and shouldn't be used to do *anything* remotely serious

## Installation
* Binary
```shell
$ go install github.com/mitander/bitrush
```
* Library
```shell
$ go get -u github.com/mitander/bitrush
```

## Usage
* Binary
```shell
$ bitrush -f <path-to-torrent-file>
```
* Library
```go
path := "example.torrent"
m, err := metainfo.NewMetaInfo(path)
if err != nil {
    log.Fatal(err)
}

t, err := torrent.NewTorrent(m)
if err != nil {
    log.Fatal(err)
}

err = t.Download()
if err != nil {
    log.Fatal(err)
}
```
## License
[MIT License](LICENSE).
