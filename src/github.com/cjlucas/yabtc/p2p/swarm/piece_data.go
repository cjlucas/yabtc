package swarm

import (
	"github.com/cjlucas/yabtc/p2p/messages"
	"github.com/cjlucas/yabtc/torrent"
)

type pieceData struct {
	piece  *torrent.Piece
	blocks []*messages.Piece
}

type pieceDataWriter struct {
	PieceDataChan chan *pieceData
	fs            *torrent.FileStream
	ErrorChan     chan error
}

func newPieceData(p *torrent.Piece) *pieceData {
	pd := &pieceData{}
	pd.piece = p
	pd.blocks = make([]*messages.Piece, 0, pd.numBlocks())
	return pd
}

func (pd *pieceData) Done() bool {
	return len(pd.blocks) == cap(pd.blocks)
}

func (pd *pieceData) bytes() []byte {
	data := make([]byte, pd.piece.Length)

	for _, block := range pd.blocks {
		copy(data[block.Begin:], block.Block)
	}

	return data
}

func (pd *pieceData) numBlocks() int {
	numBlocks := pd.piece.Length / BLOCK_SIZE
	if pd.piece.Length%BLOCK_SIZE > 0 {
		numBlocks++
	}

	return numBlocks
}

func newPieceDataWriter(fs *torrent.FileStream) *pieceDataWriter {
	return &pieceDataWriter{
		PieceDataChan: make(chan *pieceData),
		fs:            fs,
		ErrorChan:     make(chan error),
	}
}

func (w *pieceDataWriter) Write(p *pieceData) {
	go func() {
		w.PieceDataChan <- p
	}()
}

func (w *pieceDataWriter) Run() {
	for {
		pd, ok := <-w.PieceDataChan
		if !ok {
			break
		}

		b := torrent.Block{
			Offset: pd.piece.ByteOffset,
			Length: pd.piece.Length,
		}

		if err := w.fs.WriteBlock(b, pd.bytes()); err != nil {
			w.ErrorChan <- err
		}
	}
}
