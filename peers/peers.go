package peers

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

// http://www.bittorrent.org/beps/bep_0020.html
var peerIDPrefix = []byte("-RN0001-")

type PeerID [20]byte

type Peer struct {
	IP   net.IP
	Port uint16
}

func Unmarshal(bin []byte) ([]Peer, error) {
	const size = 6           // peer size - 4 for ip, 2 for port
	count := len(bin) / size // number of peers
	if len(bin)%size != 0 {
		err := fmt.Errorf("Invalid peers length")
		return nil, err
	}
	peers := make([]Peer, count)
	for i := 0; i < count; i++ {
		offset := i * size
		ip := offset + 4   // bin[offset:ip] is IP
		port := offset + 6 // bin[ip:port] is Port
		peers[i].IP = net.IP(bin[offset:ip])
		peers[i].Port = binary.BigEndian.Uint16([]byte(bin[ip:port]))
	}
	return peers, nil
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

func GeneratePeerID() (PeerID, error) {
	var id PeerID
	copy(id[:], peerIDPrefix)
	_, err := rand.Read(id[len(peerIDPrefix):])
	return id, err
}
