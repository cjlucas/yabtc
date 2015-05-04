package swarm

import (
	"errors"
	"fmt"

	"github.com/cjlucas/yabtc/bitfield"
	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/p2p/messages"
	"github.com/cjlucas/yabtc/torrent"
)

var TorrentExistsError = errors.New("torrent already exists")

type SwarmStatus int

const (
	STARTED SwarmStatus = iota
	STOPPED
)

type Stats struct {
	Downloaded int
	Uploaded   int
	Pieces     *bitfield.Bitfield
}

type Swarm struct {
	Torrent           *torrent.MetaData
	Status            SwarmStatus
	Peers             []*Peer
	peerMsgChan       chan peerMessage
	Stats             Stats
	blockReceivedChan chan *messages.Piece
}

func New(t *torrent.MetaData) *Swarm {
	s := &Swarm{}
	s.Torrent = t
	s.Status = STOPPED
	s.peerMsgChan = make(chan peerMessage, 10000)
	s.Stats.Pieces = bitfield.New(t.NumPieces())
	s.blockReceivedChan = make(chan *messages.Piece, 100)

	return s
}

func (s *Swarm) AddPeer(peer *p2p.Peer) {
	p := newPeer(peer)
	s.Peers = append(s.Peers, p)
	p.Pieces = bitfield.New(s.Torrent.NumPieces())
	p.BlockReceivedChan = s.blockReceivedChan
	go s.runPeer(p)
}

func (s *Swarm) runPeer(p *Peer) {
	// TODO defer remove from s.Peers
	p.Peer.StartHandlers()
	defer p.Peer.Disconnect()

	p.Peer.WriteChan <- messages.NewBitfield(s.Stats.Pieces)
	p.Peer.WriteChan <- messages.NewInterested()

	for {
		select {
		case msg, ok := <-p.Peer.ReadChan:
			if !ok {
				return
			}
			s.peerMsgChan <- peerMessage{p, msg}
		case <-p.Peer.ClosedConnChan:
			return
		}
	}
}

func (s *Swarm) monitorSwarm() {
	fmt.Println("monitor")
	for _, p := range s.Peers {
		if p.Choked {
			continue
		}
		for i := 0; i < s.Stats.Pieces.Length(); i++ {
			if s.Stats.Pieces.Get(i) == 0 && p.Pieces.Get(i) == 1 {
				for j := 0; j < 10; j++ {
					p.sendBlockRequest(messages.NewRequest(i, j, 1<<14))
				}
			}
		}
	}
}

func (s *Swarm) Run() {
	reqd := false
	for {
		select {
		case pmsg := <-s.peerMsgChan:
			pmsg.peer.handleMessage(pmsg.msg)
		case msg := <-s.blockReceivedChan:
			// TODO: Cancel any pending requests for received block
			fmt.Printf("received a block! %s\n", msg)
		default:
			if !reqd && len(s.Peers) > 0 && !s.Peers[0].Choked {
				s.monitorSwarm()
				reqd = true
			}
		}
	}
}
