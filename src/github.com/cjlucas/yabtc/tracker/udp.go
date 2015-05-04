package tracker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
)

const initialUdpConnectionId = 0x41727101980

type connectInput struct {
	ConnectionId  uint64
	Action        uint32 // 0
	TransactionId uint32
}

type connectOutput struct {
	Action        uint32 // 0
	TransactionId uint32
	ConnectionId  uint64
}

type announceInput struct {
	ConnectionId  uint64
	Action        uint32 // 1
	TransactionId uint32
	InfoHash      [20]byte
	PeerId        [20]byte
	Downloaded    uint64
	Left          uint64
	Uploaded      uint64
	Event         uint32
	Ip            uint32
	Key           uint32
	NumWant       int32 // -1
	Port          uint16
}

type announceOutput struct {
	Action        uint32 // 1
	TransactionId uint32
	Interval      uint32
	Leechers      uint32
	Seeders       uint32
}

type announceOutputPeer struct {
	Ip_   uint32
	Port_ uint16
}

type udpAnnounceResponse struct {
	rawAnnounce *announceOutput
	rawPeers    []*announceOutputPeer
}

func (p *announceOutputPeer) Ip() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(p.Ip_&0xff000000)>>24,
		(p.Ip_&0xff0000)>>16,
		(p.Ip_&0xff00)>>8,
		p.Ip_&0xff)
}

func (p *announceOutputPeer) Port() int {
	return int(p.Port_)
}

func (p *announceOutputPeer) PeerId() []byte {
	return nil
}

func (r *udpAnnounceResponse) FailureReason() string {
	switch r.rawAnnounce.Action {
	case 3:
		return "unknown error"
	default:
		return ""
	}
}

func (r *udpAnnounceResponse) Interval() int {
	return int(r.rawAnnounce.Interval)
}

func (r *udpAnnounceResponse) TrackerId() string {
	return ""
}

func (r *udpAnnounceResponse) Seeders() int {
	return int(r.rawAnnounce.Seeders)
}

func (r *udpAnnounceResponse) Leechers() int {
	return int(r.rawAnnounce.Leechers)
}

func (r *udpAnnounceResponse) Peers() []Peer {
	peers := make([]Peer, len(r.rawPeers))

	for i := range peers {
		peers[i] = r.rawPeers[i]
	}

	return peers
}

func eventNum(e string) int {
	switch e {
	case "started":
		return 0
	default:
		return 0
	}
}

func startUdpConnection(conn net.Conn) (uint64, error) {

	connInput := connectInput{
		ConnectionId:  initialUdpConnectionId,
		Action:        0,
		TransactionId: rand.Uint32(),
	}

	if err := binary.Write(conn, binary.BigEndian, &connInput); err != nil {
		return 0, err
	}

	var connOutput connectOutput
	if err := binary.Read(conn, binary.BigEndian, &connOutput); err != nil {
		return 0, err
	}

	if connInput.TransactionId != connOutput.TransactionId {
		return 0, fmt.Errorf("transaction id mismatch")
	}

	return connOutput.ConnectionId, nil
}

func performUdpAnnounce(conn net.Conn, connId uint64, r *AnnounceRequest) (AnnounceResponse, error) {
	in := announceInput{
		ConnectionId:  connId,
		Action:        1,
		TransactionId: rand.Uint32(),
		Downloaded:    uint64(r.Downloaded),
		Left:          uint64(r.Left),
		Uploaded:      uint64(r.Uploaded),
		Event:         uint32(eventNum(r.Event)),
		Ip:            0,
		Key:           0,
		NumWant:       -1,
		Port:          uint16(r.Port),
	}

	fmt.Printf("%02X\n", r.InfoHash)
	copy(in.InfoHash[:], r.InfoHash)
	copy(in.PeerId[:], r.PeerId)

	fmt.Printf("%#v %02X\n", in, in.InfoHash[:])

	if err := binary.Write(conn, binary.BigEndian, &in); err != nil {
		return nil, err
	}

	// Read response in buffer because successive read calls hang for some reason
	buf := make([]byte, 4096)
	if n, err := conn.Read(buf); err != nil {
		return nil, err
	} else if n < 20 {
		return nil, errors.New("not enough bytes read for announce output")
	} else {
		bytesLeft := n
		buffer := bytes.NewBuffer(buf)

		var out announceOutput
		if err := binary.Read(buffer, binary.BigEndian, &out); err != nil {
			return nil, err
		}

		bytesLeft -= 20

		fmt.Printf("%#v\n", out)
		if in.TransactionId != out.TransactionId {
			return nil, fmt.Errorf("transaction id mismatch")
		}

		var resp udpAnnounceResponse
		resp.rawAnnounce = &out
		resp.rawPeers = make([]*announceOutputPeer, bytesLeft/6)

		for i := range resp.rawPeers {
			var peer announceOutputPeer
			if err := binary.Read(buffer, binary.BigEndian, &peer); err != nil {
				return nil, err
			}
			resp.rawPeers[i] = &peer
		}

		return &resp, nil
	}
}

func udpRequest(r *AnnounceRequest) (AnnounceResponse, error) {
	var host string // host:port
	fmt.Println(r.Url)
	if u, err := url.Parse(r.Url); err != nil {
		return nil, fmt.Errorf("error parsing url: %s", err)
	} else {
		host = u.Host
	}

	fmt.Println(host)
	if conn, err := net.Dial("udp", host); err != nil {
		return nil, fmt.Errorf("could not connect: %s", err)
	} else {
		connId, err := startUdpConnection(conn)
		if err != nil {
			return nil, err
		}

		return performUdpAnnounce(conn, connId, r)
	}
}
