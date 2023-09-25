package tracker

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
	"github.com/mitander/bitrush/peers"
)

const TrackerPort = 6889

type Tracker struct {
	Announce string
	Query    string
	PeerId   [20]byte
}

func NewTracker(announce string, length int, infoHash [20]byte, peerId [20]byte) (Tracker, error) {
	u, err := url.Parse(announce)
	if err != nil {
		return Tracker{}, err
	}

	// url.Values requires map[]string
	p := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(int(TrackerPort))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(length)},
	}

	u.RawQuery = p.Encode()
	return Tracker{
		Announce: announce,
		Query:    u.String(),
		PeerId:   peerId,
	}, nil
}

type trackerRes struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (t *Tracker) ReqPeers() ([]peers.Peer, error) {
	c := &http.Client{Timeout: 15 * time.Second}
	res, err := c.Get(t.Query)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response := trackerRes{}
	err = bencode.Unmarshal(res.Body, &response)
	if err != nil {
		return nil, err
	}
	return peers.Unmarshal([]byte(response.Peers))
}
