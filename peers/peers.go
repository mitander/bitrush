package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

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
		offset := i * count
		ip := offset + 4   // bin[offset:ip] is IP
		port := offset + 6 // bin[ip:port] is Port
		peers[i].IP = net.IP(bin[offset:ip])
		peers[i].Port = binary.BigEndian.Uint16([]byte(bin[ip:port]))
	}
	return peers, nil
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.FormatInt(int64(p.Port), 64))
}
