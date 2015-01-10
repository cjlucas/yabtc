package main

import (
	"fmt"

	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/torrent"
)

import (
	"github.com/cjlucas/yabtc/services"
	"github.com/cjlucas/yabtc/services/swarm_manager"
)

import "os"

/*
 *func dunno() {
 *    t, _ := torrent.ParseFile(os.Args[1])
 *    hash := t.InfoHash()
 *    fmt.Printf("%x\n", hash)
 *
 *    resp, err := tracker.Announce(t.Announce, hash)
 *
 *    if err != nil {
 *        panic(err)
 *    }
 *
 *    peer := resp.Peers()[0]
 *
 *    conn, err := p2p.New(fmt.Sprintf("%s:%d", peer.Ip, peer.Port))
 *
 *    if err != nil {
 *        panic(err)
 *    }
 *
 *    conn.PerformHandshake(hash, hash)
 *
 *    conn.StartHandlers()
 *
 *    for {
 *        msg := <-conn.ReadChan
 *        if msg.Id == 5 {
 *            conn.WriteChan <- *messages.InterestedMessage()
 *        } else if msg.Id == 1 {
 *            conn.WriteChan <- *messages.RequestMessage(0, 0, 1024)
 *        }
 *    }
 *}
 */

func checkpiece() {
	metadata, _ := torrent.ParseFile(os.Args[1])

	t := services.NewTorrent("/Users/chris/Downloads", *metadata)

	tm := services.NewTorrentManager()

	tm.AddTorrent(t)

	tm.CheckTorrent(t)

	tm.Run()
}

func testSwarmManager() {
	metadata, _ := torrent.ParseFile(os.Args[1])
	fmt.Println(metadata.InfoHashString())

	sm := swarm_manager.NewSwarmManager()

	sm.RegisterTorrent(metadata)

	sm.VerifyPeer(metadata.InfoHashString(), p2p.NewPeer("89.85.48.189", 51413, nil))
	sm.VerifyPeer(metadata.InfoHashString(), p2p.NewPeer("95.211.141.107", 51523, nil))

	swarm_manager.Run(sm)
}

func main() {
	testSwarmManager()
}