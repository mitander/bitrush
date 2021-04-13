package handshake

//https://wiki.theory.org/BitTorrentSpecification#Handshake
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func New(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent Protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}
