package p2p

type Peer struct {
	Ip     string `bencode:"ip"`
	Port   uint32 `bencode:"port"`
	PeerId string `bencode:"peer id"`
}
