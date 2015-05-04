package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cjlucas/yabtc/torrent"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func main() {
	pm, err := NewPeerManager(54343)
	if err != nil {
		fmt.Printf("error: could not start peer manager: %s", err)
		return
	}
	go pm.Run()

	sm := NewSwarmManager()
	go sm.Run()

	tm := NewTrackerManager()

	t, _ := torrent.ParseFile(os.Args[1])
	sm.AddTorrent(t)

	peerId := []byte("-AZ2060-000000000000")
	tm.AddTracker(t.Announce, t.InfoHash(), peerId)

	pm.RegisterTorrent(t.InfoHash()[:], peerId)

	for {
		select {
		case r := <-tm.AnnounceResponseChan:
			fmt.Printf("Received tracker response: %v\n", r)
			for _, p := range r.Response.Peers() {
				pm.VerifyPeer(r.InfoHash[:], p.Ip(), p.Port())
			}
		case vp := <-pm.VerifiedPeerChan:
			fmt.Println("here", vp.Peer.Ip())
			sm.AddPeer(vp.InfoHash, vp.Peer)
		}
	}
}
