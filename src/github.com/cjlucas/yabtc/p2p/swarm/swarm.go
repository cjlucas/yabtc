package swarm

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/cjlucas/yabtc/bitfield"
	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/p2p/messages"
	"github.com/cjlucas/yabtc/torrent"
)

var TorrentExistsError = errors.New("torrent already exists")

const BLOCK_SIZE = 1 << 14

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
	Stats             Stats
	peerMessageChan   chan PeerMessage
	blockReceivedChan chan *messages.Piece
	pendingPieces     map[int]*pieceData
	pieceWriter       *pieceDataWriter
}

func requestPiece(writeChan chan<- messages.Message, p *torrent.Piece) {
	bytesLeft := p.Length
	offset := 0

	for bytesLeft > BLOCK_SIZE {
		writeChan <- messages.NewRequest(p.Index, offset, BLOCK_SIZE)
		offset += BLOCK_SIZE
		bytesLeft -= BLOCK_SIZE
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

		for i := range s.Torrent.GeneratePieces() {
			requestPiece(p.Peer.WriteChan, &s.Torrent.GeneratePieces()[i])
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

	pd.blocks = append(pd.blocks, msg)

	if pd.Done() {
		delete(s.pendingPieces, msg.Index)
		fmt.Println("HEY I RECEIVED A FULL PIECE")
		s.Stats.Pieces.Set(msg.Index, 1)
		s.pieceWriter.Write(pd)

		if msg.Index+1 == len(s.Torrent.GeneratePieces()) {
			fmt.Println(s.Stats.Pieces.Bytes())
			return
		}
	}

}

func (s *Swarm) Run() {
	reqd := false

	fs := torrent.NewFileStream("", s.Torrent.Files())
	s.pieceWriter = newPieceDataWriter(fs)

	go s.pieceWriter.Run()

	for {
		select {
		case <-s.peerMessageChan:
		case msg := <-s.blockReceivedChan:
			// TODO: Cancel any pending requests for received block
			s.handleNewBlock(msg)
		case err := <-s.pieceWriter.ErrorChan:
			fmt.Printf("Received error when writing %s\n", err)
		case <-time.NewTicker(1 * time.Second).C:
			fmt.Println(runtime.NumGoroutine())
			if !reqd && len(s.Peers) > 0 && !s.Peers[0].Choked {
				s.monitorSwarm()
				reqd = true
			}
		}
	}
}
