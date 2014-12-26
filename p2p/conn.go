package p2p

import (
	"encoding/binary"
	"fmt"
	"github.com/cjlucas/yabtc/p2p/messages"
	"io"
	"net"
)

type P2PConn struct {
	Conn      net.Conn
	ReadChan  chan messages.Message
	WriteChan chan messages.Message
}

// TODO Use Peer as input arg
func New(addr string) (*P2PConn, error) {
	var p2pConn P2PConn
	conn, err := net.Dial("tcp", addr)
	p2pConn.Conn = conn

	if err != nil {
		return nil, err
	}

	/*
	 *_, err = readHandshake(conn)
	 *if err != nil {
	 *    return nil, err
	 *}
	 */

	p2pConn.ReadChan = make(chan messages.Message)
	p2pConn.WriteChan = make(chan messages.Message)

	return &p2pConn, nil
}

func (c *P2PConn) PerformHandshake(infoHash []byte, peerId []byte) error {
	// TODO: use peerid
	hs := NewHandshake("BitTorrent protocol", infoHash, peerId)

	c.Conn.Write(hs.Bytes())
	readHandshake(c.Conn)
	return nil
}

func (c *P2PConn) StartHandlers() {
	go readHandler(c.Conn, c.ReadChan)
	go writeHandler(c.Conn, c.WriteChan)
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
