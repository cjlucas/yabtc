package tracker

import "fmt"

type Peer interface {
	Ip() string
	Port() int
	PeerId() []byte // Can be nil
}

type dictFormatPeer struct {
	IpStr     string `bencode:"ip"`
	PortInt   int    `bencode:"port"`
	PeerIdStr string `bencode:"peer id"`
}

type binaryFormatPeer struct {
	data []byte
}

func (p *dictFormatPeer) Ip() string {
	return p.IpStr
}

func (p *dictFormatPeer) Port() int {
	return p.PortInt
}

func (p *dictFormatPeer) PeerId() []byte {
	return []byte(p.PeerIdStr)
}

func (p *binaryFormatPeer) Ip() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		p.data[0], p.data[1], p.data[2], p.data[3])
}

func (p *binaryFormatPeer) Port() int {
	return (int(p.data[4]) << 8) | int(p.data[5])
}

func (p *binaryFormatPeer) PeerId() []byte {
	return nil
}
