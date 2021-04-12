package torrentfile

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
	"github.com/mitander/bitrush/peers"
)

type bencodeTrackerRes struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (tf *TorrentFile) parseTrackerUrl(peerID [20]byte, port uint16) (string, error) {
	u, err := url.Parse(tf.Announce)
	if err != nil {
		return "", err
	}

	p := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(1))}, // TODO: 1 is throwaway -  create Port struct
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Length)},
	}

	u.RawQuery = p.Encode()
	return u.String(), nil

}

func (tf *TorrentFile) reqPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := tf.parseTrackerUrl(peerID, port)
	if err != nil {
		return nil, err
	}

	c := &http.Client{Timeout: 15 * time.Second}
	res, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	tracker := bencodeTrackerRes{}
	err = bencode.Unmarshal(res.Body, &tracker)
	if err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(tracker.Peers))
}
