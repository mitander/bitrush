package p2p

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"
)

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
