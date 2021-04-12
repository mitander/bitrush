package peers

import (
	"encoding/binary"
	"fmt"
	"net"
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
		start := offset + 4
		end := offset + 6
		peers[i].IP = net.IP(bin[offset:end])
		peers[i].Port = binary.BigEndian.Uint16([]byte(bin[start:end]))
	}
	return peers, nil
}
