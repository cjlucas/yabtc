package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/cjlucas/yabtc/tracker"
)

const DEFAULT_ANNOUNCE_INTERVAL = 30 * time.Minute

const NUM_ANNOUNCE_WORKERS = 5

type AnnounceResponseInfo struct {
	InfoHash [20]byte
	Url      string
	Response tracker.AnnounceResponse
	Error    error
}

type trackerInfo struct {
	InfoHash          [20]byte
	Url               string
	PeerId            [20]byte
	nextAnnounceTimer *time.Timer
	announceQueue     chan *trackerInfo
}

type trackerInfoKey struct {
	InfoHash [20]byte
	Url      string
}

type TrackerManager struct {
	AnnounceResponseChan chan *AnnounceResponseInfo
	trackers             map[trackerInfoKey]*trackerInfo
	trackersLock         sync.RWMutex
	announceQueue        chan *trackerInfo
}

func (t *trackerInfo) setNextAnnounceTimer(d time.Duration) {
	t.nextAnnounceTimer = time.AfterFunc(d, func() {
		t.announceQueue <- t
	})
}

func (tm *TrackerManager) getTrackerInfo(url string, infoHash []byte) *trackerInfo {
	key := trackerInfoKey{Url: url}
	copy(key.InfoHash[:], infoHash)

	tm.trackersLock.RLock()
	t := tm.trackers[key]
	tm.trackersLock.RUnlock()

	return t
}

func (tm *TrackerManager) announceWorker() {
	for {
		t, ok := <-tm.announceQueue
		if !ok {
			break
		}

		// don't attempt to announce if tracker has been removed
		key := trackerInfoKey{InfoHash: t.InfoHash, Url: t.Url}
		tm.trackersLock.RLock()
		t = tm.trackers[key]
		tm.trackersLock.RUnlock()

		if t == nil {
			continue
		}

		fmt.Println("hereeeee")
		req := tracker.AnnounceRequest{
			Url:      t.Url,
			InfoHash: t.InfoHash[:],
			PeerId:   t.PeerId[:],
		}

		if resp, err := req.Request(); err != nil {
			fmt.Printf("err: %s\n", err)
			t.setNextAnnounceTimer(DEFAULT_ANNOUNCE_INTERVAL)
		} else {
			fmt.Println("here2")
			respInfo := &AnnounceResponseInfo{
				InfoHash: t.InfoHash,
				Response: resp,
				Url:      t.Url,
				Error:    err}
			tm.AnnounceResponseChan <- respInfo

			// If tracker doesnt give an announce interval,
			// be nice and wait the default interval
			nextInterval := DEFAULT_ANNOUNCE_INTERVAL
			if resp.Interval() > 0 {
				nextInterval = time.Duration(resp.Interval()) * time.Second
			}
			t.setNextAnnounceTimer(nextInterval)
		}
	}

}

// TODO: don't use metadata, use an actual torrent struct
// which contains all of the stats related to it
func (tm *TrackerManager) AddTracker(url string, infoHash []byte, peerId []byte) {
	ti := &trackerInfo{}
	copy(ti.InfoHash[:], infoHash)
	copy(ti.PeerId[:], peerId)

	ti.Url = url
	ti.setNextAnnounceTimer(0 * time.Second)

	key := trackerInfoKey{Url: url}
	ti.announceQueue = tm.announceQueue
	copy(key.InfoHash[:], infoHash)

	tm.trackersLock.Lock()
	tm.trackers[key] = ti
	tm.trackersLock.Unlock()
}

func (tm *TrackerManager) RemoveTracker(url string, infoHash []byte) {
	if t := tm.getTrackerInfo(url, infoHash); t != nil {
		t.nextAnnounceTimer.Stop()
	}

	key := trackerInfoKey{Url: url}
	copy(key.InfoHash[:], infoHash)

	tm.trackersLock.Lock()
	delete(tm.trackers, key)
	tm.trackersLock.Unlock()
}

func (tm *TrackerManager) ForceAnnounce(url string, infoHash []byte) {
	if t := tm.getTrackerInfo(url, infoHash); t != nil {
		t.nextAnnounceTimer.Reset(0)
	}
}

func (tm *TrackerManager) Stop() {
	close(tm.announceQueue)
}

func NewTrackerManager() *TrackerManager {
	tm := &TrackerManager{}
	tm.AnnounceResponseChan = make(chan *AnnounceResponseInfo)
	tm.announceQueue = make(chan *trackerInfo, 100)
	tm.trackers = make(map[trackerInfoKey]*trackerInfo)

	for i := 0; i < NUM_ANNOUNCE_WORKERS; i++ {
		go tm.announceWorker()
	}

	return tm
}

//func main() {
//tm := NewTrackerManager()

//fmt.Println("before first for")
//fmt.Printf("Got %d torrents\n", len(os.Args[1:]))
//for _, fpath := range os.Args[1:] {
//if t, err := torrent.ParseFile(fpath); err != nil {
//panic(err)
//} else {
//tm.AddTracker(t.Announce, t.InfoHash(), t.InfoHash())
//fmt.Printf("Added %s\n", fpath)
//}
//}

//fmt.Println("before 2nd for")
//i := 0
//for {
//respInfo := <-tm.AnnounceResponseChan
//i++
//fmt.Printf("received %d responses\n", i)
//fmt.Println(respInfo.Url)
//fmt.Println(respInfo.Response.FailureReason)
/*
 *for _, p := range respInfo.Response.Peers() {
 *    fmt.Printf("%s:%d\n", p.Ip(), p.Port())
 *}
 */
//}
//}
