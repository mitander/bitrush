package torrentfile

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mitander/bitrush/peers"
	"github.com/stretchr/testify/assert"
)

func TestParseTrackerUrl(t *testing.T) {
	const port uint16 = 6889
	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	tf := TorrentFile{
		Announce: "http://test.tracker.org:6969/announce",
		InfoHash: [20]byte{148, 102, 213, 85, 174, 246, 146, 126, 127, 246, 85, 15, 22, 6, 186, 128, 220, 105, 12, 15},
		PieceHashes: [][20]byte{
			{84, 48, 101, 49, 83, 50, 116, 51, 80, 52, 105, 53, 69, 54, 99, 55, 69, 56, 115, 57},
			{84, 48, 101, 49, 83, 50, 116, 51, 80, 52, 105, 53, 69, 54, 99, 55, 69, 56, 115, 57},
		},
		PieceLength: 262144,
		Length:      351272960,
		Name:        "test.iso",
	}
	url, err := tf.parseTrackerUrl(peerID, port)
	expected := "http://test.tracker.org:6969/announce?compact=1&downloaded=0&info_hash=%94f%D5U%AE%F6%92~%7F%F6U%0F%16%06%BA%80%DCi%0C%0F&left=351272960&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6889&uploaded=0"
	assert.Nil(t, err)
	assert.Equal(t, url, expected)
}

func TestRequestPeers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []byte(
			"d" +
				"8:interval" + "i900e" +
				"5:peers" + "12:" +
				string([]byte{
					192, 0, 2, 210, 0x1A, 0xE8, // 0x1AE8 = 6888
					127, 0, 0, 21, 0x1A, 0xE9, // 0x1AE9 = 6889
				}) + "e")
		w.Write(response)
	}))
	defer ts.Close()
	tf := TorrentFile{
		Announce: ts.URL,
		InfoHash: [20]byte{148, 102, 213, 85, 174, 246, 146, 126, 127, 246, 85, 15, 22, 6, 186, 128, 220, 105, 12, 15},
		PieceHashes: [][20]byte{
			{84, 48, 101, 49, 83, 50, 116, 51, 80, 52, 105, 53, 69, 54, 99, 55, 69, 56, 115, 57},
			{84, 48, 101, 49, 83, 50, 116, 51, 80, 52, 105, 53, 69, 54, 99, 55, 69, 56, 115, 57},
		},
		PieceLength: 262144,
		Length:      351272960,
		Name:        "test.iso",
	}
	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	const port uint16 = 6882
	expected := []peers.Peer{
		{IP: net.IP{192, 0, 2, 210}, Port: 6888},
		{IP: net.IP{127, 0, 0, 21}, Port: 6889},
	}
	p, err := tf.ReqPeers(peerID, port)
	assert.Nil(t, err)
	assert.Equal(t, expected, p)
}
