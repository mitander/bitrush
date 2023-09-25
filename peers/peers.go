package peers

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// http://www.bittorrent.org/beps/bep_0020.html
var peerIDPrefix = []byte("-BR0001-")

type PeerID [20]byte

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

func Unmarshal(b []byte) ([]Peer, error) {
	// 4 bytes ip, 2 bytes port
	const size = 6
	count := len(b) / size
	if len(b)%size != 0 {
		err := errors.New("invalid compact peer list length")
		log.Error(err.Error())
		return nil, err
	}
	peers := make([]Peer, count)
	for i := 0; i < count; i++ {
		offset := i * size
		ip := offset + 4
		port := offset + 6

		// IP: b[offset:ip] Port: b[ip:port]
		peers[i].IP = net.IP(b[offset:ip])
		peers[i].Port = binary.BigEndian.Uint16([]byte(b[ip:port]))
	}
	return peers, nil
}

func GeneratePeerID() (PeerID, error) {
	var id PeerID
	copy(id[:], peerIDPrefix)
	_, err := rand.Read(id[len(peerIDPrefix):])
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("failed to generate peer id")
		return PeerID{}, err
	}
	return id, nil
}
