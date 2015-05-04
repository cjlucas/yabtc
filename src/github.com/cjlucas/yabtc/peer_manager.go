package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/cjlucas/yabtc/p2p"
)

type HandshakeInfo struct {
	InfoHash [20]byte
	PeerId   [20]byte
}

type HandshakeInfoRequest struct {
	InfoHash [20]byte
	C        chan *HandshakeInfo
}

type VerifiedPeer struct {
	InfoHash, PeerId []byte
	Peer             *p2p.Peer
}

type PeerManager struct {
	VerifiedPeerChan     chan VerifiedPeer
	ln                   net.Listener
	Infos                map[[20]byte]*HandshakeInfo
	registerTorrentChan  chan *HandshakeInfo
	handshakeInfoReqChan chan *HandshakeInfoRequest
}

func NewPeerManager(port int) (*PeerManager, error) {
	m := &PeerManager{}

	if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
		return nil, err
	} else {
		m.ln = ln
	}

	m.Infos = make(map[[20]byte]*HandshakeInfo)
	m.handshakeInfoReqChan = make(chan *HandshakeInfoRequest)
	m.registerTorrentChan = make(chan *HandshakeInfo)
	m.VerifiedPeerChan = make(chan VerifiedPeer)

	return m, nil
}

func (m *PeerManager) RegisterTorrent(infoHash, peerId []byte) {
	hi := &HandshakeInfo{}
	copy(hi.InfoHash[:], infoHash)
	copy(hi.PeerId[:], peerId)

	m.registerTorrentChan <- hi
}

func (m *PeerManager) getHandshakeInfo(infoHash []byte) *HandshakeInfo {
	c := make(chan *HandshakeInfo)
	var hash [20]byte
	copy(hash[:], infoHash)
	m.handshakeInfoReqChan <- &HandshakeInfoRequest{hash, c}
	return <-c
}

func (m *PeerManager) recvHandshake(peer *p2p.Peer) (*p2p.Handshake, error) {
	hsIn, err := peer.ReceiveHandshake()
	if err != nil {
		logger.Printf("error recving handshake (%s:%d): %s", peer.Ip(), peer.Port(), err)
		peer.Disconnect()
		return nil, err
	}

	handshakeInfo := m.getHandshakeInfo(hsIn.InfoHash[:])
	if handshakeInfo == nil {
		return nil, errors.New("received peer handshaking with unknown info hash")
	}

	return hsIn, nil
}

func (m *PeerManager) sendHandshake(peer *p2p.Peer, infoHash []byte) error {
	handshakeInfo := m.getHandshakeInfo(infoHash)
	hs := p2p.NewHandshake("BitTorrent protocol", handshakeInfo.InfoHash[:], handshakeInfo.PeerId[:])
	if err := peer.SendHandshake(*hs); err != nil {
		logger.Printf("error sending handshake (%s:%d): %s", peer.Ip(), peer.Port(), err)
		return err
	}

	return nil
}

func (m *PeerManager) verifyPeer(peer *p2p.Peer, infoHash []byte) {
	logger.Printf("Verifying peer %s:%d", peer.Ip(), peer.Port())
	if peer.IsConnected() {
		if hs, err := m.recvHandshake(peer); err != nil {
			return
		} else if err = m.sendHandshake(peer, hs.InfoHash[:]); err != nil {
			return
		} else {
			m.VerifiedPeerChan <- VerifiedPeer{hs.InfoHash[:], hs.PeerId[:], peer}
		}
	} else {
		if err := peer.Connect(); err != nil {
			return
		}

		if err := m.sendHandshake(peer, infoHash); err != nil {
			return
		}

		if hs, err := m.recvHandshake(peer); err != nil {
			return
		} else {
			m.VerifiedPeerChan <- VerifiedPeer{hs.InfoHash[:], hs.PeerId[:], peer}
		}
	}
}

func (m *PeerManager) VerifyPeer(infoHash []byte, ip string, port int) {
	peer := p2p.NewPeer(ip, port)
	go m.verifyPeer(peer, infoHash)
}

func (m *PeerManager) runPeerListener() {
	for {
		conn, err := m.ln.Accept()
		if err != nil {
			return
		}
		peer := p2p.NewPeerWithConn(conn)
		go m.verifyPeer(peer, nil)
	}
}

func (m *PeerManager) Run() {
	go m.runPeerListener()

	for {
		select {
		case info := <-m.registerTorrentChan:
			m.Infos[info.InfoHash] = info
		case req := <-m.handshakeInfoReqChan:
			req.C <- m.Infos[req.InfoHash]
		}
	}
}
