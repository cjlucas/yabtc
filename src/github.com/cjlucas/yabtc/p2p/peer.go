package p2p

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/cjlucas/yabtc/p2p/messages"
	"github.com/cjlucas/yabtc/piece"
)

const READ_DEADLINE = 1 * time.Second

type Peer struct {
	Addr          PeerAddr
	peerId        [20]byte
	Conn          net.Conn
	Pieces        []piece.Piece
	Choked        bool
	Interested    bool
	BytesReceived int
	BytesSent     int
	ReadChan      chan messages.Message
	quitReadChan  chan bool
	WriteChan     chan messages.Message
	quitWriteChan chan bool
}

type PeerAddr struct {
	Ip   string
	Port int
}

func (addr PeerAddr) Network() string {
	return "tcp"
}

func (addr PeerAddr) String() string {
	return fmt.Sprintf("%s:%d", addr.Ip, addr.Port)
}

func NewPeer(ip string, port int, peerId [20]byte, pieces []piece.Piece) *Peer {
	var peer Peer

	peer.Addr = PeerAddr{ip, port}
	peer.peerId = peerId
	peer.Pieces = pieces

	peer.ReadChan = make(chan messages.Message, 100)
	peer.quitReadChan = make(chan bool)
	peer.WriteChan = make(chan messages.Message, 100)
	peer.quitWriteChan = make(chan bool)

	return &peer
}

func (p Peer) Ip() string {
	return p.Addr.Ip
}

func (p Peer) Port() int {
	return p.Addr.Port
}

func (p Peer) HasPeerId() bool {
	for _, b := range p.peerId {
		if b != 0 {
			return true
		}
	}
	return false
}

func (p Peer) PeerId() [20]byte {
	return p.peerId
}

func (p *Peer) IsConnected() bool {
	return p.Conn != nil
}

func (p *Peer) Connect() error {
	if conn, err := net.Dial(p.Addr.Network(), p.Addr.String()); err != nil {
		return err
	} else {
		p.Conn = conn
	}
	return nil

}

func (p *Peer) Disconnect() {
	p.quitReadChan <- true
	p.quitWriteChan <- true

	<-p.quitReadChan
	<-p.quitWriteChan

	if p.IsConnected() {
		p.Conn.Close()
		p.Conn = nil
	}
}

func (p *Peer) SendHandshake(hs Handshake) error {
	_, err := p.Conn.Write(hs.Bytes())
	return err
}

func (p *Peer) ReceiveHandshake(expectedInfoHash [20]byte) error {
	if hs_resp, err := readHandshake(p.Conn); err != nil {
		return err
	} else if p.HasPeerId() && hs_resp.PeerId == p.PeerId() {
		return fmt.Errorf("peer ID does not match")
	} else if hs_resp.InfoHash != expectedInfoHash {
		return fmt.Errorf("info hash does not match")
	}

	return nil
}

func (p *Peer) StartHandlers() {
	go p.readHandler()
	go p.writeHandler()
}

func readBytes(r io.Reader, buf []byte, count int) error {
	bytesRead := 0
	for bytesRead < count {
		cnt, err := r.Read(buf[bytesRead:])
		if err != nil {
			return err
		}

		bytesRead += cnt
	}

	return nil
}

func readHandshake(r io.Reader) (*Handshake, error) {
	var resp Handshake
	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	resp.Plen = int(buf[0])

	remainingHandshakeBytes := resp.Plen + 48
	buf = make([]byte, remainingHandshakeBytes)
	if err := readBytes(r, buf, remainingHandshakeBytes); err != nil {
		return nil, err
	}

	resp.Pstr = string(buf[0:resp.Plen])
	copy(resp.Reserved[0:], buf[resp.Plen:resp.Plen+8])
	copy(resp.InfoHash[0:], buf[resp.Plen+8:resp.Plen+28])
	copy(resp.PeerId[0:], buf[resp.Plen+28:resp.Plen+48])

	return &resp, nil
}

func readMessage(r net.Conn) (*messages.Message, error) {
	r.SetReadDeadline(time.Now().Add(READ_DEADLINE))

	var msg messages.Message
	if err := binary.Read(r, binary.BigEndian, &msg.Len); err != nil {
		return nil, err
	}

	if msg.Len > 0 {
		buf := make([]byte, msg.Len)
		if err := readBytes(r, buf, len(buf)); err != nil {
			return nil, err
		}

		msg.Id = buf[0]
		msg.Payload = buf[1:]
	}

	return &msg, nil
}

func (p *Peer) readHandler() {
	for {
		select {
		case <-p.quitReadChan:
			p.quitReadChan <- true
			return
		default:
			// TODO handle error
			msg, err := readMessage(p.Conn)

			if err != nil {
				continue
			}

			p.ReadChan <- *msg
		}
	}
}

func (p *Peer) writeHandler() {
	for {
		select {
		case msg := <-p.WriteChan:
			p.Conn.Write(msg.Bytes())
		case <-p.quitWriteChan:
			p.quitWriteChan <- true
			return
		}
	}
}
