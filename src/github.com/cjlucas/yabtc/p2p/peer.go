package p2p

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/cjlucas/yabtc/p2p/messages"
)

const READ_DEADLINE = 5 * time.Second

type Peer struct {
	Addr           PeerAddr
	peerId         [20]byte
	Conn           net.Conn
	Choked         bool
	Interested     bool
	BytesReceived  int
	BytesSent      int
	ReadChan       chan messages.Message
	WriteChan      chan messages.Message
	ClosedConnChan chan bool
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

func NewPeer(ip string, port int) *Peer {
	var peer Peer

	peer.Addr = PeerAddr{ip, port}

	peer.ReadChan = make(chan messages.Message, 100)
	peer.WriteChan = make(chan messages.Message, 100)
	peer.ClosedConnChan = make(chan bool)

	return &peer
}

func NewPeerWithConn(conn net.Conn) *Peer {
	var ip string
	var port int
	fmt.Sscanf(conn.RemoteAddr().String(), "%s:%d", ip, port)
	p := NewPeer(ip, port)
	p.Conn = conn
	return p
}

func (p Peer) Ip() string {
	return p.Addr.Ip
}

func (p Peer) Port() int {
	return p.Addr.Port
}

func (p Peer) Address() string {
	return fmt.Sprintf("%s:%d", p.Ip(), p.Port())
}

func (p Peer) PeerId() [20]byte {
	return p.peerId
}

func (p *Peer) IsConnected() bool {
	return p.Conn != nil
}

func (p *Peer) Connect() error {
	dialer := net.Dialer{READ_DEADLINE, time.Time{}, nil, true, 0}
	if conn, err := dialer.Dial("tcp", p.Address()); err != nil {
		return err
	} else {
		p.Conn = conn
	}
	return nil

}

func (p *Peer) Disconnect() {
	if p.IsConnected() {
		p.Conn.Close()
		p.Conn = nil
	}
}

func (p *Peer) SendHandshake(hs Handshake) error {
	return writeBytes(p.Conn, hs.Bytes())
}

func (p *Peer) ReceiveHandshake() (*Handshake, error) {
	if hs_resp, err := readHandshake(p.Conn); err != nil {
		return nil, err
	} else {
		return hs_resp, nil
	}
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

func writeBytes(w io.Writer, buf []byte) error {
	bytesWritten := 0
	for bytesWritten < len(buf) {
		if n, err := w.Write(buf[bytesWritten:]); err != nil {
			return err
		} else {
			bytesWritten += n
		}
	}

	return nil
}

func readHandshake(r net.Conn) (*Handshake, error) {
	r.SetReadDeadline(time.Now().Add(READ_DEADLINE))

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

func readMessage(r net.Conn) (messages.Message, error) {
	r.SetReadDeadline(time.Now().Add(READ_DEADLINE))

	var msgLen uint32
	if err := binary.Read(r, binary.BigEndian, &msgLen); err != nil {
		return nil, err
	}

	if msgLen > 0 {
		buf := make([]byte, 4+msgLen)
		binary.BigEndian.PutUint32(buf[0:4], msgLen)
		if err := readBytes(r, buf[4:], len(buf[4:])); err != nil {
			return nil, err
		}

		return messages.ParseBytes(buf)
	} else {
		return nil, fmt.Errorf("received message with invalid length: %d", msgLen)
	}
}

func (p *Peer) readHandler() {
	for {
		if msg, err := readMessage(p.Conn); err == nil {
			p.ReadChan <- msg
		} else if err == io.EOF {
			p.ClosedConnChan <- true
			return
		} else if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
			fmt.Println("readMessage error ", err)
		}
	}
}

func (p *Peer) writeHandler() {
	for {
		select {
		case msg := <-p.WriteChan:
			//fmt.Println("will write", msg)
			if err := writeBytes(p.Conn, messages.AsBytes(msg)); err == io.EOF {
				p.ClosedConnChan <- true
				return
			} else if err != nil {
				fmt.Println("writeeBytes error ", err)
			}
		}
	}
}
