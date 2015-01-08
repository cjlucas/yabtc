package p2p

import "github.com/cjlucas/yabtc/torrent"

type Swarm struct {
	TorInfo     torrent.MetaData
	LocalPeerId []byte
	Peers       []Peer
}

func NewSwarm(torInfo torrent.MetaData) *Swarm {
	var s Swarm
	s.TorInfo = torInfo
	s.LocalPeerId = generateLocalPeerId()
	return &s
}

func (s *Swarm) AddPeer(peer *Peer) {
	s.Peers = append(s.Peers, *peer)
}

func generateLocalPeerId() []byte {
	return []byte("cheese is so so good")
}
