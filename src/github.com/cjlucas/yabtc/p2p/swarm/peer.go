package swarm

import (
	"fmt"

	"github.com/cjlucas/yabtc/bitfield"
	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/p2p/messages"
)

type peerMessage struct {
	peer *Peer
	msg  messages.Message
}

type Peer struct {
	Peer *p2p.Peer

	// Are we choked by peer?
	Choked bool

	// Is peer interested in us?
	Interested bool

	Pieces *bitfield.Bitfield

	// Pending incoming block requests
	InBlockRequests []*messages.Request

	// Pending outgoing block requests
	OutBlockRequests []*messages.Request

	// Chan to notify Swarm of an incoming block
	BlockReceivedChan chan<- *messages.Piece
}

func (p *Peer) Ip() string {
	return p.Peer.Ip()
}

func (p *Peer) Port() int {
	return p.Peer.Port()
}

func newPeer(peer *p2p.Peer) *Peer {
	p := &Peer{Choked: true, Interested: false, Peer: peer}
	p.InBlockRequests = make([]*messages.Request, 0)
	p.OutBlockRequests = make([]*messages.Request, 0)
	return p
}

func (p *Peer) sendBlockRequest(msg *messages.Request) bool {
	for _, pendingMsg := range p.OutBlockRequests {
		if *pendingMsg == *msg {
			return false
		}
	}

	p.OutBlockRequests = append(p.OutBlockRequests, msg)
	p.Peer.WriteChan <- msg
	return true
}

func (p *Peer) handleMessage(msg messages.Message) {
	fmt.Printf("handlePeerMessage %s:%d\n", p.Ip(), p.Port())
	fmt.Println(msg)

	switch msg := msg.(type) {
	case *messages.Bitfield:
		p.Pieces.SetBytes(msg.Bits.Bytes())
	case *messages.Interested:
		p.Interested = true
	case *messages.NotInterested:
		p.Interested = false
		p.InBlockRequests = make([]*messages.Request, 0)
	case *messages.Choke:
		p.Choked = true
	case *messages.Unchoke:
		p.Choked = false
	case *messages.Have:
		p.Pieces.Set(msg.PieceIndex, 1)
		// TODO: scan incoming block requests, remove if matching block found
	case *messages.Request:
		p.InBlockRequests = append(p.InBlockRequests, msg)
	case *messages.Piece:
		//p.BlockReceivedChan <- msg
	default:
		fmt.Println("got unknown message")
	}
}
