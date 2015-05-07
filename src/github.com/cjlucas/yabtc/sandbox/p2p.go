package main

import (
	"fmt"
	"os"

	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/p2p/messages"
	"github.com/cjlucas/yabtc/torrent"
)

const BLOCK_MAX = 1 << 14

func blocksForPiece(p torrent.FullPiece) []torrent.Block {
	blocks := make([]torrent.Block, p.Length/BLOCK_MAX)

	curOffset := 0
	bytesLeft := p.Length
	for i := 0; i < len(blocks); i++ {
		b := &blocks[i]
		b.Offset = curOffset

		if bytesLeft < BLOCK_MAX {
			b.Length = bytesLeft
		} else {
			b.Length = BLOCK_MAX
		}

		curOffset += b.Length
		bytesLeft -= b.Length
	}

	return blocks
}

func reqPiece(p *p2p.Peer, curPiece int, blocks []torrent.Block) {
	for _, b := range blocks {
		fmt.Println(b)
		req := messages.NewRequest(curPiece, b.Offset, b.Length)
		p.WriteChan <- req
	}
}

func main() {

	fmt.Println(os.Args[1])
	t, err := torrent.ParseFile(os.Args[1])
	fmt.Println("after parse")
	fmt.Println(t.InfoHashString())

	if err != nil {
		panic(err)
	}

	//p := p2p.NewPeer("0.0.0.0", 51413)
	p := p2p.NewPeer("192.168.1.19", 33144)
	fmt.Println("before connect")
	if err := p.Connect(); err != nil {
		panic(err)
	}
	fmt.Println("after connect")
	defer p.Disconnect()

	fmt.Println("before handshake")

	ihash := t.InfoHash()
	/*
	 *ihash[0] = 0
	 */
	hs := p2p.NewHandshake("BitTorrent protocol", ihash, []byte("-AZ2060-000000000000"))
	if err := p.SendHandshake(*hs); err != nil {
		panic(err)
	}

	fmt.Println("after send handshake")

	var hash [20]byte
	copy(hash[:], t.InfoHash())
	fmt.Println("before recv handshake")

	if hs_resp, err := p.ReceiveHandshake(); err != nil {
		panic(err)
	} else {
		fmt.Println(hs_resp)
	}

	fmt.Println("after recv handshake")

	p.StartHandlers()
	p.WriteChan <- messages.NewInterested()
	pieces := t.GeneratePieces()
	curPiece := 0
	blocks := blocksForPiece(pieces[0])
	curBlock := 0
	for {
		select {
		case msg := <-p.ReadChan:
			switch msg := msg.(type) {
			case *messages.Unchoke:
				reqPiece(p, curPiece, blocks)
			case *messages.Piece:
				fmt.Println("received: ", msg.Index, msg.Begin)
				curBlock++
				// received a full piece
				if curBlock >= len(blocks) {
					p.WriteChan <- messages.NewHave(curPiece)
					curBlock = 0
					curPiece++
					if curPiece == len(pieces) {
						fmt.Println("All done")
						p.Disconnect()
						return
					}
					blocks = blocksForPiece(pieces[curPiece])
					reqPiece(p, curPiece, blocks)
				}
			default:
				fmt.Println("received unhandled message")
			}

		case <-p.ClosedConnChan:
			fmt.Println("Connection closed")
			return
		}
	}
}
