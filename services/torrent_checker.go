package services

import (
	"bytes"
	"github.com/cjlucas/yabtc/torrent"
)

type TorrentChecker struct {
	Torrent   Torrent
	PieceChan chan torrent.FullPiece
}

func (c *TorrentChecker) Check(t *Torrent) {
	fs := torrent.FileStream{t.Root, t.MetaData.Files()}

	for _, p := range t.MetaData.GeneratePieces() {
		checksum := fs.CalculatePieceChecksum(p)

		p.Have = bytes.Equal(checksum, p.Hash)
		c.PieceChan <- p
	}

	close(c.PieceChan)
}
