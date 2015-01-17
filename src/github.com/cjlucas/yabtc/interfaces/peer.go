package interfaces

type Peer interface {
	Ip() string
	Port() int
	HasPeerId() bool
	PeerId() [20]byte
}
