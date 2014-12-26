package tracker

import (
	"fmt"
	"github.com/cjlucas/yabtc/p2p"
	"github.com/zeebo/bencode"
	"net/http"
	"net/url"
)

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
	var peers []p2p.Peer
	if err := bencode.DecodeBytes(rawPeers, &peers); err != nil {
		panic(err)
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
		p := &peers[i]

		ipBytes := rawPeers[curByte : curByte+4]
		p.Ip = fmt.Sprintf("%d.%d.%d.%d",
			ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
		p.Port = uint32(rawPeers[curByte+4])
		p.Port = p.Port << 8
		p.Port |= uint32(rawPeers[curByte+5])
		curByte += 6
	}

	return peers
}
