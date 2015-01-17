package tracker

type peer struct {
	ip     string
	port   int
	peerId [20]byte
}

func (p peer) Ip() string {
	return p.ip
}

func (p peer) Port() int {
	return p.port
}

func (p peer) HasPeerId() bool {
	for _, b := range p.peerId {
		if b != 0 {
			return true
		}
	}

	return false
}

func (p peer) PeerId() [20]byte {
	return p.peerId
}
