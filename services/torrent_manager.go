package services

import (
	"fmt"
	"github.com/cjlucas/yabtc/torrent"
	"time"
)

type TorrentStatus int

const (
	STARTED TorrentStatus = iota
	STOPPED
	CHECKING
)

type Torrent struct {
	Id               string
	MetaData         torrent.MetaData
	InfoHash         string
	Root             string
	Files            torrent.FileList
	Pieces           []byte
	Status           TorrentStatus
	NextAnnounce     time.Time
	AmountUploaded   int
	AmountDownloaded int
	AmountLeft       int
}

type TorrentManager struct {
	Torrents           map[string]*Torrent
	TorrentCheckerChan chan torrent.FullPiece
}

func NewTorrent(root string, metadata torrent.MetaData) *Torrent {
	var t Torrent
	t.Root = root
	t.MetaData = metadata
	t.Files = metadata.Files()
	t.InfoHash = metadata.InfoHashString()
	t.Id = t.InfoHash
	return &t
}

func NewTorrentManager() *TorrentManager {
	var tm TorrentManager
	tm.TorrentCheckerChan = make(chan torrent.FullPiece, 100)

	return &tm
}

func (m *TorrentManager) AddTorrent(t *Torrent) {
}

func (m *TorrentManager) DeleteTorrent(t *Torrent) {
	delete(m.Torrents, t.InfoHash)

	// TODO stop any jobs working on this torrent
}

func (m *TorrentManager) StartTorrent(t *Torrent) {

}

func (m *TorrentManager) StopTorrent(t *Torrent) {
	t.Status = STOPPED

	// TODO disconnect any peer connections
}

func (m *TorrentManager) CheckTorrent(t *Torrent) {
	var c TorrentChecker
	c.PieceChan = m.TorrentCheckerChan
	c.Torrent = *t

	go c.Check(t)
}

func (m *TorrentManager) Run() {
	for {
		select {
		case piece, ok := <-m.TorrentCheckerChan:
			if !ok {
				return
			}
			fmt.Printf("Received Piece: %d %s %s\n", piece.Index, piece.Have, ok)
		}

	}
}
