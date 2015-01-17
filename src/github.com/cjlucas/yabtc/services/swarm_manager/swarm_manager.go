package swarm_manager

import (
	"fmt"
	"time"

	"github.com/cjlucas/yabtc/interfaces"
	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/p2p/messages"
	"github.com/cjlucas/yabtc/torrent"
)

type SwarmManager struct {
	peerVerificationChan       chan peerVerification
	peerVerificationResultChan chan peerVerificationResult
	Swarms                     map[[20]byte]*p2p.Swarm
}

type peerVerification struct {
	InfoHash [20]byte
	Peer     *p2p.Peer
}

type peerVerificationResult struct {
	InfoHash [20]byte
	Peer     *p2p.Peer
	Err      error
}

func NewSwarmManager() *SwarmManager {
	var m SwarmManager
	m.peerVerificationChan = make(chan peerVerification, 1024)
	m.peerVerificationResultChan = make(chan peerVerificationResult, 1024)
	m.Swarms = make(map[[20]byte]*p2p.Swarm)

	return &m
}

func (m *SwarmManager) RegisterTorrent(torInfo *torrent.MetaData) {
	s := p2p.NewSwarm(*torInfo)
	m.Swarms[torInfo.InfoHash()] = s
}

func (m *SwarmManager) VerifyPeer(infoHash [20]byte, p *p2p.Peer) {
	result := peerVerificationResult{infoHash, p, nil}
	defer func() {
		m.peerVerificationResultChan <- result
	}()

	swarm, ok := m.Swarms[infoHash]

	if !ok {
		result.Err = fmt.Errorf("Swarm with info hash %x not found\n", infoHash)
		return
	}

	p.Pieces = swarm.EmptyPieces()

	if err := p.Connect(); err != nil {
		result.Err = err
		return
	}

	defer p.Disconnect()

	hs := p2p.NewHandshake("BitTorrent protocol", infoHash[:], swarm.LocalPeerId)
	if err := p.SendHandshake(*hs); err != nil {
		result.Err = fmt.Errorf("error sending handshake: %s\n", err)
		return
	}

	if err := p.ReceiveHandshake(infoHash); err != nil {
		result.Err = fmt.Errorf("error receiving handshake: %s\n", err)
		return
	}

	p.StartHandlers()

	select {
	case msg := <-p.ReadChan:
		if msg.Id == messages.BITFIELD_MSG_ID {
			messages.DecodeBitfieldPayload(msg.Payload, p.Pieces)

		}
	case <-time.Tick(1 * time.Second):
		//Assume empty bitfield if bitfield not received
	}

	have := 0
	for _, p := range p.Pieces {
		if p.Have {
			have++
		}
	}
}

func (m *SwarmManager) AddPeerToSwarm(infoHash [20]byte, p interfaces.Peer) {
	peer := p2p.NewPeer(p.Ip(), p.Port(), p.PeerId(), nil)
	m.peerVerificationChan <- peerVerification{infoHash, peer}
}

func (m *SwarmManager) Run() {
	for {
		select {
		case resp := <-m.peerVerificationChan:
			go m.VerifyPeer(resp.InfoHash, resp.Peer)
		case result := <-m.peerVerificationResultChan:
			if result.Err == nil {
				s := m.Swarms[result.InfoHash]
				s.AddPeer(*result.Peer)
			} else {
				fmt.Printf("Got an error: %s\n", result.Err)
			}
		}
	}
}
