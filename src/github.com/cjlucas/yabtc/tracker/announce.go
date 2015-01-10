package tracker

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/cjlucas/yabtc/p2p"
	"github.com/zeebo/bencode"
)

type peer struct {
	Ip     string `bencode:"ip"`
	Port   uint32 `bencode:"port"`
	PeerId string `bencode:"peer id"`
}

type AnnounceResponse struct {
	FailureReason  string             `bencode:"failure reason"`
	WarningMessage string             `bencode:"warning message"`
	Interval       int                `bencode:"interval"`
	MinInterval    int                `bencode:"min interval"`
	TrackerId      string             `bencode:"tracker id"`
	Complete       int                `bencode:"complete"`
	Incomplete     int                `bencode:"incomplete"`
	RawPeers       bencode.RawMessage `bencode:"peers"`
}

func (resp *AnnounceResponse) Peers() []p2p.Peer {
	switch resp.RawPeers[0] {
	case 'l':
		return parsePeersDictFormat(resp.RawPeers)
	default:
		return parsePeersBinaryFormat(resp.RawPeers)
	}
}

func Announce(announceUrl string, infoHash []byte) (*AnnounceResponse, error) {
	vals := make(url.Values)

	vals.Add("info_hash", string(infoHash))
	vals.Add("peer_id", string(infoHash))
	vals.Add("port", "9999")
	vals.Add("uploaded", "0")
	vals.Add("downloaded", "0")
	vals.Add("left", "0")
	vals.Add("event", "started")
	vals.Add("compact", "1")

	fullUrl := fmt.Sprintf("%s?%s", announceUrl, vals.Encode())

	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	decoder := bencode.NewDecoder(resp.Body)
	var out AnnounceResponse
	if err := decoder.Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

func parsePeersDictFormat(rawPeers []byte) []p2p.Peer {
	var peerList []peer
	if err := bencode.DecodeBytes(rawPeers, peerList); err != nil {
		panic(err)
	}

	peers := make([]p2p.Peer, len(peerList))

	for i, p := range peerList {
		peer := p2p.NewPeer(p.Ip, int(p.Port), []byte(p.PeerId))
		peers[i] = *peer
	}

	return peers
}

func parsePeersBinaryFormat(rawPeers []byte) []p2p.Peer {
	peers := make([]p2p.Peer, len(rawPeers)/6)

	curByte := 0
	for rawPeers[curByte] != ':' {
		curByte += 1
	}
	curByte += 1
	for i, _ := range peers {
		ipBytes := rawPeers[curByte : curByte+4]
		ip := fmt.Sprintf("%d.%d.%d.%d",
			ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])

		port := uint32(rawPeers[curByte+4])
		port = port << 8
		port |= uint32(rawPeers[curByte+5])

		peer := p2p.NewPeer(ip, int(port), nil)
		peers[i] = *peer

		curByte += 6
	}

	return peers
}