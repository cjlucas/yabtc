package tracker

import (
	"fmt"
	"net/url"
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

type AnnounceResponse interface {
	FailureReason() string
	Interval() int
	TrackerId() string
	Seeders() int
	Leechers() int
	Peers() []Peer
}

func (r *AnnounceRequest) Request() (AnnounceResponse, error) {
	if url, err := url.Parse(r.Url); err != nil {
		return nil, err
	} else {
		switch url.Scheme {
		case "http":
			fallthrough
		case "https":
			return httpRequest(r)
		case "udp":
			return udpRequest(r)
		default:
			return nil, fmt.Errorf("unknown url scheme: %s", url.Scheme)
		}
	}
}
