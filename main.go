package main

import "github.com/cjlucas/yabtc/torrent"
import "github.com/cjlucas/yabtc/tracker"
import "github.com/cjlucas/yabtc/p2p/messages"
import "github.com/cjlucas/yabtc/p2p"
import "fmt"
import "os"
import "bytes"

func dunno() {
	t, _ := torrent.ParseFile(os.Args[1])
	hash := t.InfoHash()
	fmt.Printf("%x\n", hash)

	resp, err := tracker.Announce(t.MetaInfo.Announce, hash)

	if err != nil {
		panic(err)
	}

	peer := resp.Peers()[0]

	conn, err := p2p.New(fmt.Sprintf("%s:%d", peer.Ip, peer.Port))

	if err != nil {
		panic(err)
	}

	conn.PerformHandshake(hash, hash)

	conn.StartHandlers()

	for {
		msg := <-conn.ReadChan
		if msg.Id == 5 {
			conn.WriteChan <- *messages.InterestedMessage()
		} else if msg.Id == 1 {
			conn.WriteChan <- *messages.RequestMessage(0, 0, 1024)
		}
	}
}

func checkpiece() {
	t, _ := torrent.ParseFile(os.Args[1])
	hash := t.InfoHash()
	fmt.Printf("%x\n", hash)

	fs := torrent.FileStream{"/Users/chris/Downloads", t.Files()}

	for _, p := range t.Pieces {
		if checksum := fs.CalculatePieceChecksum(p); bytes.Equal(checksum, p.Hash) {
			fmt.Printf("PASSED\n")
		} else {
			fmt.Printf("FAILED %x\n", p.Hash)
		}
	}
}

func main() {
	checkpiece()
}
