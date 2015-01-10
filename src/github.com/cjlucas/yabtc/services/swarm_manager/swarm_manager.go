package swarm_manager

import (
	"fmt"

	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/piece"
	"github.com/cjlucas/yabtc/torrent"
)

type SwarmManager struct {
	PeerVerificationChan chan PeerVerification
	Swarms               map[string]p2p.Swarm
}

type PeerVerification struct {
	InfoHash string
	Peer     *p2p.Peer
	Error    error
}

func NewSwarmManager() *SwarmManager {
	var m SwarmManager
	m.PeerVerificationChan = make(chan PeerVerification, 1024)
	m.Swarms = make(map[string]p2p.Swarm)

	return &m
}

func (m *SwarmManager) RegisterTorrent(torInfo *torrent.MetaData) {
	s := p2p.NewSwarm(*torInfo)
	m.Swarms[torInfo.InfoHashString()] = *s
}

func (m *SwarmManager) VerifyPeer(infoHash string, p *p2p.Peer) {
	go func() {
		s := m.Swarms[infoHash]
		p.Pieces = make([]piece.Piece, s.TorInfo.NumPieces())
		err := p.PerformHandshake(s.TorInfo.InfoHash(), s.LocalPeerId)
		p.Disconnect()

		pv := PeerVerification{s.TorInfo.InfoHashString(), p, err}
		m.PeerVerificationChan <- pv
	}()
}

func (m *SwarmManager) AddPeerToSwarm(infoHash string, p *p2p.Peer) {
	s := m.Swarms[infoHash]
	s.AddPeer(p)
}

func Run(m *SwarmManager) {
	for {
		select {
		case resp := <-m.PeerVerificationChan:
			if resp.Error != nil {
				fmt.Printf("Error verifying peer. Won't add to swarm. (Error: %s)\n", resp.Error)
			} else {
				fmt.Println("Adding peer")
				m.AddPeerToSwarm(resp.InfoHash, resp.Peer)

				has := 0
				for _, p := range resp.Peer.Pieces {
					if p.Have {
						has++
					}
				}

				fmt.Printf("Peer has %d of %d pieces\n", has, len(resp.Peer.Pieces))
			}
		}
	}
}
