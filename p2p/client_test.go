package p2p

import (
	"encoding/binary"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := map[string]struct {
		peer   Peer
		id     messageID
		msgLen int
		fails  bool
	}{
		"correct input": {
			peer:   Peer{IP: net.IP{127, 0, 0, 1}, Port: 1442},
			id:     MsgBitfield, // <- when creating client first msg should be bitfield
			msgLen: 3,
			fails:  false,
		},
		"invalid message id": {
			peer:   Peer{IP: net.IP{127, 0, 0, 1}, Port: 1442},
			id:     MsgHave, // <- fails here
			msgLen: 3,
			fails:  true,
		},
		"invalid message length": {
			peer:   Peer{IP: net.IP{127, 0, 0, 1}, Port: 1442},
			id:     MsgBitfield,
			msgLen: 0, // <- fails here, len 0 is keep alive
			fails:  true,
		},
	}

	for name, test := range tests {
		hash := make([]byte, 20)
		id := make([]byte, 20)
		bitfield := Bitfield{0b11111111, 0b11111111}

		// sync test with listener
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			client, err := NewClient(test.peer, [20]byte(hash), [20]byte(id))
			if test.fails {
				assert.Error(t, err, name)

			} else {
				assert.Nil(t, err, name)
				assert.Equal(t, bitfield, client.Bitfield)
			}
			wg.Done()
		}()

		// listen to connection
		l, err := net.Listen("tcp4", test.peer.String())
		assert.Nil(t, err, name)
		c, err := l.Accept()
		assert.Nil(t, err, name)

		// write handshake
		hsLen := 48
		pstrlen := make([]byte, 1)
		pstrlen[0] = byte(2)
		c.Write(pstrlen)
		hs := make([]byte, 2+hsLen)
		hs[1] = byte(MsgBitfield)
		c.Write(hs)

		// write message length
		msglen := make([]byte, 4)
		binary.BigEndian.PutUint32(msglen, uint32(test.msgLen))
		c.Write(msglen)

		// write message
		msg := make([]byte, 3)
		copy(msg[:1], []byte{byte(test.id)})
		copy(msg[1:3], bitfield)
		c.Write(msg)

		wg.Wait()
		c.Close()
		l.Close()
	}
}
