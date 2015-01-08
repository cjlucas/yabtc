package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/cjlucas/yabtc/p2p"
	"github.com/cjlucas/yabtc/p2p/messages"
	"github.com/cjlucas/yabtc/piece"
)

type Peer struct {
	Addr          net.Addr
	PeerId        []byte
	Conn          net.Conn
	Pieces        []piece.Piece
	Choked        bool
	Interested    bool
	BytesReceived int
	BytesSent     int
	ReadChan      chan messages.Message
	WriteChan     chan messages.Message
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

func readHandshake(r io.Reader) (*p2p.Handshake, error) {
	var resp p2p.Handshake
	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		panic(err)
	}

	resp.Plen = int(buf[0])
	fmt.Printf("readHandshake: plen = %d\n", resp.Plen)

	remainingHandshakeBytes := resp.Plen + 48
	buf = make([]byte, remainingHandshakeBytes)
	if err := readBytes(r, buf, remainingHandshakeBytes); err != nil {
		return nil, err
	}

	resp.Pstr = string(buf[0:resp.Plen])
	fmt.Println(copy(resp.Reserved[0:], buf[resp.Plen:resp.Plen+8]))
	fmt.Println(copy(resp.InfoHash[0:], buf[resp.Plen+8:resp.Plen+28]))
	fmt.Println(copy(resp.PeerId[0:], buf[resp.Plen+28:resp.Plen+48]))

	return &resp, nil
}

func readMessage(r io.Reader) (*messages.Message, error) {
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

func readHandler(conn net.Conn, c chan messages.Message) {
	for {
		// TODO handle error
		msg, err := readMessage(conn)

		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("Received message: (len: %d, id: %d)\n", msg.Len, msg.Id)
		c <- *msg
	}
}

func writeHandler(conn net.Conn, c chan messages.Message) {
	for {
		msg := <-c
		n, _ := conn.Write(msg.Bytes())
		fmt.Printf("Wrote %d bytes\n", n)
	}
}

func (p *Peer) PerformHandshake(infoHash []byte, peerId []byte) error {
	if err := p.Connect(); err != nil {
		return err
	}

	hs := p2p.NewHandshake("BitTorrent protocol", infoHash, peerId)
	if _, err := p.Conn.Write(hs.Bytes()); err != nil {
		return err
	}

	if hs_resp, err := readHandshake(p.Conn); err != nil {
		return err
	} else if !bytes.Equal(hs_resp.PeerId[:], p.PeerId) {
		return fmt.Errorf("peer ID does not match")
	}

	msg, err := readMessage(p.Conn)
	if err != nil {
		return err
	}

	if msg.Id != messages.BITFIELD_MSG_ID {
		return fmt.Errorf("received unexpected message: %d", msg.Id)
	}

	messages.DecodeBitfieldPayload(msg.Payload, p.Pieces)
	p.Disconnect()

	return nil
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
	if p.IsConnected() {
		p.Conn.Close()
		p.Conn = nil
	}
}
