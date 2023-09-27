package tracker

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mitander/bitrush/p2p"
	"github.com/stretchr/testify/assert"
)

func TestNewTrackerUrl(t *testing.T) {
	announce := "http://test.tracker.org:6969/announce"
	length := 351272960
	infoHash := [20]byte{148, 102, 213, 85, 174, 246, 146, 126, 127, 246, 85, 15, 22, 6, 186, 128, 220, 105, 12, 15}
	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	tr, err := NewTracker(announce, length, infoHash, peerID)
	assert.Equal(t, err, nil)
	expected := "http://test.tracker.org:6969/announce?compact=1&downloaded=0&info_hash=%94f%D5U%AE%F6%92~%7F%F6U%0F%16%06%BA%80%DCi%0C%0F&left=351272960&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6889&uploaded=0"
	assert.Equal(t, tr.Query, expected)
}

func TestRequestPeers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []byte(
			"d" + "8:interval" + "i900e" + "5:peers" + "12:" +
				string([]byte{
					192, 0, 2, 210, 0x1A, 0xE8, // 0x1AE8 = 6888
					127, 0, 0, 21, 0x1A, 0xE9, // 0x1AE9 = 6889
				}) + "e")
		w.Write(response)
	}))
	defer srv.Close()

	announce := srv.URL
	infoHash := [20]byte{148, 102, 213, 85, 174, 246, 146, 126, 127, 246, 85, 15, 22, 6, 186, 128, 220, 105, 12, 15}
	length := 351272960
	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	expected := []p2p.Peer{
		{IP: net.IP{192, 0, 2, 210}, Port: 6888},
		{IP: net.IP{127, 0, 0, 21}, Port: 6889},
	}
	tr, err := NewTracker(announce, length, infoHash, peerID)
	assert.Nil(t, err)
	p, err := tr.ReqPeers()
	assert.Nil(t, err)
	assert.Equal(t, expected, p)
}
