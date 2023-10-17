package tracker

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
	"github.com/mitander/bitrush/peer"
	log "github.com/sirupsen/logrus"
)

const TrackerPort = 6889

type Tracker struct {
	Announce string
	Query    string
	PeerId   [20]byte
	InfoHash [20]byte
	Peers    []peer.Peer
}

func NewTracker(announce string, length int, infoHash [20]byte, peerID [20]byte) (Tracker, error) {
	u, err := url.Parse(announce)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to create tracker")
		return Tracker{}, err
	}

	p := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{string(peerID[:])},
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
		PeerId:   peerID,
		InfoHash: infoHash,
	}, nil
}

type bencodeResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (t *Tracker) RequestPeers() ([]peer.Peer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.Query, nil)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to create request")
		return nil, err
	}

	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to send request")
		return nil, err
	}
	defer res.Body.Close()

	response := bencodeResponse{}
	err = bencode.Unmarshal(res.Body, &response)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to unmarshal bencode")
		return nil, err
	}
	return peer.Unmarshal([]byte(response.Peers))
}
