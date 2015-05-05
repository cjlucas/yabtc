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

const PIECE_SIZE = 1 << 14

type SwarmStatus int

const (
	STARTED SwarmStatus = iota
	STOPPED
)

type pieceData struct {
	piece        *torrent.Piece
	data         []byte
	bytesWritten *bitfield.Bitfield
}

func newPieceData(p *torrent.Piece) *pieceData {
	pd := &pieceData{}
	pd.piece = p
	pd.data = make([]byte, p.Length)
	pd.bytesWritten = bitfield.New(p.Length)
	return pd
}

func (pd *pieceData) write(data []byte, offset int) {
	copy(pd.data[offset:], data)

	for i := 0; i < len(data); i++ {
		pd.bytesWritten.Set(i+offset, 1)
	}
}

type Stats struct {
	Downloaded int
	Uploaded   int
	Pieces     *bitfield.Bitfield
}

type Swarm struct {
	Torrent           *torrent.MetaData
	Status            SwarmStatus
	Peers             []*Peer
	Stats             Stats
	peerMessageChan   chan PeerMessage
	blockReceivedChan chan *messages.Piece
	pendingPieces     map[int]*pieceData
}

func requestPiece(writeChan chan<- messages.Message, p *torrent.Piece) {
	bytesLeft := p.Length
	offset := 0

	for bytesLeft > PIECE_SIZE {
		writeChan <- messages.NewRequest(p.Index, offset, PIECE_SIZE)
		offset += PIECE_SIZE
		bytesLeft -= PIECE_SIZE
	}

	if bytesLeft > 0 {
		writeChan <- messages.NewRequest(p.Index, offset, bytesLeft)
	}
}

func New(t *torrent.MetaData) *Swarm {
	s := &Swarm{}
	s.Torrent = t
	s.Status = STOPPED
	s.peerMessageChan = make(chan PeerMessage, 10000)
	s.Stats.Pieces = bitfield.New(t.NumPieces())
	s.blockReceivedChan = make(chan *messages.Piece, 100)
	s.pendingPieces = make(map[int]*pieceData)

	return s
}

func (s *Swarm) PiecesSeen() []int {
	pieces := make([]int, s.Torrent.NumPieces())
	for _, p := range s.Peers {
		for i := 0; i < p.Pieces.Length(); i++ {
			if p.Pieces.Get(i) == 1 {
				pieces[i]++
			}
		}
	}

	return pieces
}

func (s *Swarm) AddPeer(peer *p2p.Peer) {
	p := newPeer(peer)
	s.Peers = append(s.Peers, p)
	p.Pieces = bitfield.New(s.Torrent.NumPieces())
	p.BlockReceivedChan = s.blockReceivedChan
	p.PeerMessageChan = s.peerMessageChan
	go p.Run()

	p.Peer.WriteChan <- messages.NewBitfield(s.Stats.Pieces)
	p.Peer.WriteChan <- messages.NewInterested()
}

func (s *Swarm) monitorSwarm() {
	fmt.Println("monitor")
	for _, p := range s.Peers {
		if p.Choked {
			continue
		}

		for i := 0; i < s.Stats.Pieces.Length(); i++ {
			if s.Stats.Pieces.Get(i) == 0 && p.Pieces.Get(i) == 1 {
				requestPiece(p.Peer.WriteChan, &s.Torrent.GeneratePieces()[i])
			}
		}
	}
}

func (s *Swarm) handleNewBlock(msg *messages.Piece) {
	pd, ok := s.pendingPieces[msg.Index]

	if pd == nil || !ok {
		p := s.Torrent.GeneratePieces()[msg.Index]
		pd = newPieceData(&p)
		s.pendingPieces[msg.Index] = pd
	}

	pd.write(msg.Block, msg.Begin)

	// TODO: this is sloooow. Replace pd.bytesWritten with a bitfield of blocks received
	// where each piece is about 250000 bytes per piece, there are only 15 blocks per piece
	for i := 0; i < pd.bytesWritten.Length(); i++ {
		if pd.bytesWritten.Get(i) == 0 {
			return
		}
	}

	s.pendingPieces[msg.Index] = nil
	fmt.Println("HEY I RECEIVED A FULL PIECE")
}

func (s *Swarm) Run() {
	reqd := false
	for {
		select {
		case <-s.peerMessageChan:
		case msg := <-s.blockReceivedChan:
			// TODO: Cancel any pending requests for received block
			s.handleNewBlock(msg)
		default:
			if !reqd && len(s.Peers) > 0 && !s.Peers[0].Choked {
				s.monitorSwarm()
				reqd = true
			}
		}
	}
}
