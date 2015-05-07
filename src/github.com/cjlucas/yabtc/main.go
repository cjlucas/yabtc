package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/cjlucas/yabtc/torrent"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

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
