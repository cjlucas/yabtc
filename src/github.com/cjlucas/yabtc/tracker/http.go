package tracker

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/zeebo/bencode"
)

type httpAnnounceResponse struct {
	FailureReason_  string             `bencode:"failure reason"`
	WarningMessage_ string             `bencode:"warning message"`
	Interval_       int                `bencode:"interval"`
	MinInterval_    int                `bencode:"min interval"`
	TrackerId_      string             `bencode:"tracker id"`
	Complete_       int                `bencode:"complete"`
	Incomplete_     int                `bencode:"incomplete"`
	Peers_          bencode.RawMessage `bencode:"peers"`
}

func (r *httpAnnounceResponse) FailureReason() string {
	return r.FailureReason_
}

func (r *httpAnnounceResponse) WarningMessage() string {
	return r.WarningMessage_
}

func (r *httpAnnounceResponse) Interval() int {
	return r.Interval_
}

func (r *httpAnnounceResponse) MinInterval() int {
	return r.MinInterval_
}

func (r *httpAnnounceResponse) TrackerId() string {
	return r.TrackerId_
}

func (r *httpAnnounceResponse) Seeders() int {
	return r.Complete_
}

func (r *httpAnnounceResponse) Leechers() int {
	return r.Incomplete_
}

func (resp *httpAnnounceResponse) Peers() []Peer {
	if resp.Peers_ == nil {
		return nil
	}

	switch resp.Peers_[0] {
	case 'l':
		return parsePeersDictFormat(resp.Peers_)
	default:
		return parsePeersBinaryFormat(resp.Peers_)
	}
}

func (r *AnnounceRequest) announceUrl() (string, error) {
	if r.Url == "" {
		return "", errors.New("invalid Url value")
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
	//vals.Add("port", fmt.Sprintf("%d", r.Port))
	vals.Add("port", fmt.Sprintf("%d", 9999))
	vals.Add("uploaded", fmt.Sprintf("%d", r.Uploaded))
	vals.Add("downloaded", fmt.Sprintf("%d", r.Downloaded))
	//vals.Add("left", fmt.Sprintf("%d", r.Left))
	vals.Add("left", fmt.Sprintf("%d", 1))
	vals.Add("event", r.Event)
	vals.Add("compact", "1")

	return fmt.Sprintf("%s?%s", r.Url, vals.Encode()), nil
}

func httpRequest(r *AnnounceRequest) (AnnounceResponse, error) {
	if announceUrl, err := r.announceUrl(); err != nil {
		return nil, fmt.Errorf("error building url string: %s", err)
	} else {
		fmt.Println(announceUrl)
		client := http.Client{}
		if req, err := http.NewRequest("GET", announceUrl, nil); err != nil {
			return nil, fmt.Errorf("new request error: %s", err)
		} else {
			req.Header.Add("User-Agent", "Transmission/2.11")
			resp, err := client.Do(req)

			if err != nil {
				return nil, fmt.Errorf("HTTP error: %s", err)
			}

			if resp.StatusCode != 200 {
				return nil, fmt.Errorf("unexpected HTTP status code: %s", resp.Status)
			}

			defer resp.Body.Close()
			decoder := bencode.NewDecoder(resp.Body)
			var out httpAnnounceResponse
			if err := decoder.Decode(&out); err != nil {
				return nil, fmt.Errorf("bencode decoding error: %s", err)
			}

			return &out, nil
		}
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
