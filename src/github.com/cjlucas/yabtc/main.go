package main

import (
	"fmt"

	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/torrent"
)

import (
	"github.com/cjlucas/yabtc/services"
	"github.com/cjlucas/yabtc/services/swarm_manager"
	"github.com/cjlucas/yabtc/services/tracker_manager"
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

	p := p2p.NewPeer("89.85.48.189", 51413, [20]byte{}, nil)
	for i := 0; i < 25; i++ {
		sm.AddPeerToSwarm(metadata.InfoHash(), p)
	}

	sm.Run()
}

func testTrackerManager() {
	t := tracker_manager.New()

	metadata, _ := torrent.ParseFile(os.Args[1])
	fmt.Println(metadata.InfoHashString())

	go t.Run()

	t.RegisterTorrent(metadata.InfoHash(), []string{metadata.Announce})

	for {
		select {
		case t := <-t.TrackerResponseChan:
			for _, p := range t.LastResponse.Peers() {
				fmt.Printf("%+v\n", p)
			}
		}
	}
}

func testTorrentManager() {
	t := services.NewTorrentManager()

	metadata, _ := torrent.ParseFile(os.Args[1])
	t.AddTorrent(services.NewTorrent("/Users/chris", *metadata))

	t.Run()
}

func main() {
	testTorrentManager()
}
