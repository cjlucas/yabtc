package tracker_manager

import (
	"fmt"
	"time"

	"github.com/cjlucas/yabtc/tracker"
)

type TrackerManager struct {
	trackers              []Tracker
	addTrackerChan        chan Tracker
	UnregisterTorrentChan chan [20]byte
	updateTrackerChan     chan *Tracker
	TrackerResponseChan   chan Tracker
}

type Tracker struct {
	InfoHash     [20]byte
	Url          string
	LastUpdated  time.Time
	NextUpdate   time.Time
	LastResponse tracker.AnnounceResponse
}

func New() *TrackerManager {
	var m TrackerManager
	m.addTrackerChan = make(chan Tracker, 1000)
	m.UnregisterTorrentChan = make(chan [20]byte, 1000)
	m.updateTrackerChan = make(chan *Tracker, 1000)
	m.TrackerResponseChan = make(chan Tracker, 1000)

	return &m
}

func (m *TrackerManager) RegisterTorrent(infoHash [20]byte, trackerUrls []string) {

	for _, url := range trackerUrls {
		fmt.Printf("Registering tracker: %s\n", url)
		var t Tracker
		t.InfoHash = infoHash
		t.Url = url

		m.addTrackerChan <- t
	}

}

func (m *TrackerManager) UnregisterTorrent(infoHash [20]byte) {
	m.UnregisterTorrentChan <- infoHash
}

func (m *TrackerManager) Run() {
	for {
		select {
		case t := <-m.addTrackerChan:
			m.trackers = append(m.trackers, t)
		case infoHash := <-m.UnregisterTorrentChan:
			m.trackers = make([]Tracker, 0)
			for _, t := range m.trackers {
				if t.InfoHash != infoHash {
					m.trackers = append(m.trackers, t)
				}
			}
		case t := <-m.updateTrackerChan:
			fmt.Printf("Fetching %s...\n", t.Url)
			resp, err := tracker.Announce(t.Url, t.InfoHash[:])
			t.LastResponse = *resp

			if err != nil {
				fmt.Printf("Error updating tracker %s (%s)\n", t.Url, err)
				break
			} else {
				fmt.Printf("Got %d Peers\n", len(resp.Peers()))
				t.LastUpdated = time.Now()
				nextUpdateDuration := time.Duration(resp.Interval) * time.Second
				t.NextUpdate = t.LastUpdated.Add(nextUpdateDuration)
				m.TrackerResponseChan <- *t
			}

		default:
			now := time.Now()

			for i := range m.trackers {
				t := &m.trackers[i]
				if now.After(t.NextUpdate) {
					m.updateTrackerChan <- t
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
}
