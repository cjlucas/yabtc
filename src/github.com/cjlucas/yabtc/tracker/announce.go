package tracker

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/zeebo/bencode"
)

type AnnounceRequest struct {
	Url        string
	InfoHash   []byte
	PeerId     []byte
	Port       int
	Uploaded   int
	Downloaded int
	Left       int
	Event      string
	NumWant    int
	//Key        string
	//TrackerId  string
}

type AnnounceResponse struct {
	FailureReason  string `bencode:"failure reason"`
	WarningMessage string `bencode:"warning message"`
	Interval       int    `bencode:"interval"`
	MinInterval    int    `bencode:"min interval"`
	TrackerId      string `bencode:"tracker id"`
	Complete       int    `bencode:"complete"`
	Incomplete     int    `bencode:"incomplete"`
	RawPeers       []byte `bencode:"peers"`
}

func (resp *AnnounceResponse) Peers() []Peer {
	if resp.RawPeers == nil {
		return nil
	}
	switch resp.RawPeers[0] {
	case 'l':
		return parsePeersDictFormat(resp.RawPeers)
	default:
		return parsePeersBinaryFormat(resp.RawPeers)
	}
}

func (r *AnnounceRequest) announceUrl() (string, error) {
	if r.Url == "" {
		return "", errors.New("TrackerUrl field is an empty string")
	}

	if len(r.InfoHash) != 20 {
		return "", errors.New("invalid InfoHash value")
	}

	if len(r.PeerId) != 20 {
		return "", errors.New("invalid PeerId value")
	}

	vals := make(url.Values)
	vals.Add("info_hash", string(r.InfoHash))
	vals.Add("peer_id", string(r.PeerId))
	vals.Add("port", fmt.Sprintf("%d", r.Port))
	vals.Add("uploaded", fmt.Sprintf("%d", r.Uploaded))
	vals.Add("downloaded", fmt.Sprintf("%d", r.Downloaded))
	vals.Add("left", fmt.Sprintf("%d", r.Left))
	vals.Add("event", r.Event)
	vals.Add("compact", "1")

	return fmt.Sprintf("%s?%s", r.Url, vals.Encode()), nil
}

func (r *AnnounceRequest) Request() (*AnnounceResponse, error) {
	if announceUrl, err := r.announceUrl(); err != nil {
		return nil, err
	} else {
		resp, err := http.Get(announceUrl)
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
}

func parsePeersDictFormat(rawPeers []byte) []Peer {
	var peerList []dictFormatPeer
	if err := bencode.DecodeBytes(rawPeers, &peerList); err != nil {
		panic(err)
	}

	peers := make([]Peer, len(peerList))
	for i := range peerList {
		peers[i] = &peerList[i]
	}

	return peers
}

func parsePeersBinaryFormat(rawPeers []byte) []Peer {
	peers := make([]Peer, len(rawPeers)/6)

	for i := range peers {
		peers[i] = &binaryFormatPeer{rawPeers[i*6 : (i*6)+6]}
	}

	return peers
}
