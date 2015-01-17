package services

import (
	"fmt"

	"github.com/cjlucas/yabtc/services/swarm_manager"
	"github.com/cjlucas/yabtc/services/tracker_manager"
	"github.com/cjlucas/yabtc/torrent"
)

type TorrentStatus int

const (
	STARTED TorrentStatus = iota
	STOPPED
	CHECKING
)

type Torrent struct {
	InfoHash         [20]byte
	MetaData         torrent.MetaData
	Root             string
	Files            torrent.FileList
	Pieces           []byte
	Status           TorrentStatus
	AmountUploaded   int
	AmountDownloaded int
	AmountLeft       int
}

type TorrentManager struct {
	Torrents           map[string]*Torrent
	SwarmManager       *swarm_manager.SwarmManager
	TrackerManager     *tracker_manager.TrackerManager
	TorrentCheckerChan chan torrent.FullPiece
}

func NewTorrent(root string, metadata torrent.MetaData) *Torrent {
	var t Torrent
	t.Root = root
	t.MetaData = metadata
	t.Files = metadata.Files()
	t.InfoHash = metadata.InfoHash()
	return &t
}

func NewTorrentManager() *TorrentManager {
	var tm TorrentManager
	tm.TorrentCheckerChan = make(chan torrent.FullPiece, 100)

	tm.SwarmManager = swarm_manager.NewSwarmManager()
	tm.TrackerManager = tracker_manager.New()

	return &tm
}

func (m *TorrentManager) AddTorrent(t *Torrent) {
	m.SwarmManager.RegisterTorrent(&t.MetaData)
	m.TrackerManager.RegisterTorrent(t.InfoHash, []string{t.MetaData.Announce})
}

func (m *TorrentManager) DeleteTorrent(t *Torrent) {

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
	fmt.Println("Starting services")
	go m.SwarmManager.Run()
	go m.TrackerManager.Run()

	fmt.Println("Services started")

	for {
		select {
		case piece, ok := <-m.TorrentCheckerChan:
			if !ok {
				return
			}
			fmt.Printf("Received Piece: %d %s %s\n", piece.Index, piece.Have, ok)
		case t := <-m.TrackerManager.TrackerResponseChan:
			resp := t.LastResponse
			for _, p := range resp.Peers() {
				m.SwarmManager.AddPeerToSwarm(t.InfoHash, p)
			}
		}
	}
}
