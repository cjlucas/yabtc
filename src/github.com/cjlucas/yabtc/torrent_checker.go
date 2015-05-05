package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/cjlucas/yabtc/torrent"
)

type TorrentCheckerProgress struct {
	Pieces []bool // TODO: use bitfield.Bitfield instead
}

type TorrentChecker struct {
}

func check(fs *torrent.FileStream, pieces []torrent.Piece, progChan chan *TorrentCheckerProgress, quit chan bool) {
	defer close(progChan)
	defer close(quit)

	progress := TorrentCheckerProgress{}

	curPiece := 0
	for {
		select {
		case <-quit:
			return
		default:
			if curPiece >= len(pieces) {
				return
			}

			p := &pieces[curPiece]
			checksum := fs.CalculatePieceChecksum(torrent.Block{p.ByteOffset, p.Length})

			progress.Pieces = append(progress.Pieces, bytes.Equal(checksum, p.Hash))

			progChan <- &progress
			curPiece++
		}
	}
}

func (c *TorrentChecker) Check(root string, metadata *torrent.MetaData) (chan *TorrentCheckerProgress, chan bool) {
	fs := torrent.FileStream{root, metadata.Files()}

	progChan := make(chan *TorrentCheckerProgress)
	quit := make(chan bool)

	go check(&fs, metadata.GeneratePieces(), progChan, quit)

	return progChan, quit
}

func main() {
	t, err := torrent.ParseFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(t.Info.Name)
	fmt.Println(t.Files())

	tc := TorrentChecker{}
	ch, quit := tc.Check("/Users/chris/Downloads", t)
	time.AfterFunc(1*time.Millisecond, func() {
		quit <- true
	})
	for progress := range ch {
		trues := 0
		for _, b := range progress.Pieces {
			if b {
				trues++
			}
		}
		fmt.Println(len(progress.Pieces))
	}

	fmt.Println("end")
}
