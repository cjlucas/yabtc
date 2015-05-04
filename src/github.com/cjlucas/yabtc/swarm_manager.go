package main

import (
	"sync"

	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/p2p/swarm"
	"github.com/cjlucas/yabtc/torrent"
)

type SwarmManager struct {
	Swarms         map[[20]byte]*swarm.Swarm
	swarmLock      sync.RWMutex
	addTorrentChan chan *torrent.MetaData
}

func NewSwarmManager() *SwarmManager {
	m := &SwarmManager{}

	m.Swarms = make(map[[20]byte]*swarm.Swarm)
	m.addTorrentChan = make(chan *torrent.MetaData)

	return m
}

// Assumes torrent with given info hash is not already addded
func (m *SwarmManager) AddTorrent(t *torrent.MetaData) {
	m.addTorrentChan <- t
}

func (m *SwarmManager) AddPeer(infoHash []byte, peer *p2p.Peer) {
	var hash [20]byte
	copy(hash[:], infoHash)
	s := m.Swarms[hash]

	s.AddPeer(peer)
}

func (m *SwarmManager) handleNewSwarm(t *torrent.MetaData) {
	logger.Printf("Adding new swarm for torrent: %d", t.InfoHash())
	s := swarm.New(t)

	var hash [20]byte
	copy(hash[:], t.InfoHash())
	m.Swarms[hash] = s
	go s.Run()
}

func (m *SwarmManager) Run() {
	for {
		select {
		case t := <-m.addTorrentChan:
			m.handleNewSwarm(t)
		}
	}
}
