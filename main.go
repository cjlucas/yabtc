package main

import "github.com/cjlucas/yabtc/torrent"
import "github.com/cjlucas/yabtc/tracker"
import "github.com/cjlucas/yabtc/p2p/messages"
import "github.com/cjlucas/yabtc/p2p"
import "fmt"
import "os"

func main() {
	tp, _ := torrent.ParseFile(os.Args[1])
	hash := tp.MetaInfo.InfoHash()
	fmt.Printf("%x\n", hash)

	resp, err := tracker.Announce(tp.MetaInfo.Announce, hash)

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
