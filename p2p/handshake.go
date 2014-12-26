package p2p

type Handshake struct {
	Plen     int
	Pstr     string
	Reserved [8]byte
	InfoHash [20]byte
	PeerId   [20]byte
}

func NewHandshake(pstr string, infoHash []byte, peerId []byte) *Handshake {
	var handshake Handshake
	handshake.Plen = len(pstr)
	handshake.Pstr = pstr
	for i := 0; i < len(handshake.Reserved); i++ {
		handshake.Reserved[i] = 0
	}
	copy(handshake.InfoHash[0:], infoHash)
	copy(handshake.PeerId[0:], peerId)

	return &handshake
}

func (h *Handshake) Bytes() []byte {
	buf := make([]byte, h.Plen+49)
	offset := 0
	buf[offset] = byte(h.Plen)
	offset += 1

	copy(buf[offset:], h.Pstr[0:])
	offset += len(h.Pstr)

	copy(buf[offset:], h.Reserved[0:])
	offset += len(h.Reserved)

	copy(buf[offset:], h.InfoHash[0:])
	offset += len(h.InfoHash)

	copy(buf[offset:], h.PeerId[0:])
	offset += len(h.PeerId)

	return buf
}
