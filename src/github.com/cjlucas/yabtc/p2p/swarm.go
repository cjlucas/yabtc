package p2p

import (
	"fmt"

	"github.com/cjlucas/yabtc/piece"
	"github.com/cjlucas/yabtc/torrent"
)

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

func (s *Swarm) AddPeer(peer Peer) {
	s.Peers = append(s.Peers, peer)

	fmt.Printf("I have %d peers\n", len(s.Peers))
}

func (s *Swarm) EmptyPieces() []piece.Piece {
	pieces := make([]piece.Piece, s.TorInfo.NumPieces())

	for i := 0; i < len(pieces); i++ {
		pieces[i].Offset = i
		pieces[i].Have = false
		pieces[i].Size = s.TorInfo.PieceSize()
	}

	return pieces
}

func generateLocalPeerId() []byte {
	return []byte("cheese is so so good")
}
