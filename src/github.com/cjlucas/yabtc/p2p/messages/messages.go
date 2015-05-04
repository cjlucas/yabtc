package messages

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cjlucas/yabtc/bitfield"
)

const (
	CHOKE_MSG_ID          = 0
	UNCHOKE_MSG_ID        = 1
	INTERESTED_MSG_ID     = 2
	NOT_INTERESTED_MSG_ID = 3
	HAVE_MSG_ID           = 4
	BITFIELD_MSG_ID       = 5
	REQUEST_MSG_ID        = 6
	PIECE_MSG_ID          = 7
	CANCEL_MSG_ID         = 8
	PORT_MSG_ID           = 9
)

type Message interface {
	Id() int
	Payload() []byte
}

type Generic struct {
	id      int
	payload []byte
}

type Choke struct{}

type Unchoke struct{}

type Interested struct{}

type NotInterested struct{}

type Have struct {
	PieceIndex int
}

type Bitfield struct {
	Bits *bitfield.Bitfield
}

type Request struct {
	Index, Begin, Length int
}

type Piece struct {
	Index, Begin int
	Block        []byte
}

type Cancel struct {
	Index, Begin, Length int
}

type Port struct {
	Port int
}

func AsBytes(m Message) []byte {
	buf := make([]byte, 4+1+len(m.Payload()))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(buf)-4))
	buf[4] = byte(m.Id())
	copy(buf[5:], m.Payload())

	return buf
}

func WriteTo(m Message, w io.Writer) error {
	buf := AsBytes(m)

	bytesWritten := 0
	for bytesWritten < len(buf) {
		n, err := w.Write(buf[bytesWritten:])
		if err != nil {
			return err
		}

		bytesWritten += n
	}

	return nil
}

func ParseBytes(bytes []byte) (Message, error) {
	msgId := bytes[4]
	payload := bytes[5:]
	switch msgId {
	case CHOKE_MSG_ID:
		return NewChoke(), nil
	case UNCHOKE_MSG_ID:
		return NewUnchoke(), nil
	case INTERESTED_MSG_ID:
		return NewInterested(), nil
	case NOT_INTERESTED_MSG_ID:
		return NewNotInterested(), nil
	case HAVE_MSG_ID:
		msg := &Have{}
		return msg, msg.decodePayload(payload)
	case BITFIELD_MSG_ID:
		msg := &Bitfield{}
		return msg, msg.decodePayload(payload)
	case REQUEST_MSG_ID:
		msg := &Request{}
		return msg, msg.decodePayload(payload)
	case PIECE_MSG_ID:
		msg := &Piece{}
		return msg, msg.decodePayload(payload)
	case CANCEL_MSG_ID:
		msg := &Cancel{}
		return msg, msg.decodePayload(payload)
	case PORT_MSG_ID:
		msg := &Port{}
		return msg, msg.decodePayload(payload)
	default:
		return &Generic{int(msgId), payload}, nil
	}
}

func NewChoke() *Choke {
	return &Choke{}
}

func NewUnchoke() *Unchoke {
	return &Unchoke{}
}

func NewInterested() *Interested {
	return &Interested{}
}

func NewNotInterested() *NotInterested {
	return &NotInterested{}
}

func NewHave(pieceIndex int) *Have {
	return &Have{pieceIndex}
}

func NewBitfield(bits *bitfield.Bitfield) *Bitfield {
	return &Bitfield{bits}
}

func NewRequest(index, begin, length int) *Request {
	return &Request{index, begin, length}
}

func NewPiece(index, begin int, block []byte) *Piece {
	return &Piece{index, begin, block}
}

func NewCancel(index, begin, length int) *Cancel {
	return &Cancel{index, begin, length}
}

func NewPort(port int) *Port {
	return &Port{port}
}

func (m *Generic) Id() int       { return m.id }
func (m *Choke) Id() int         { return CHOKE_MSG_ID }
func (m *Unchoke) Id() int       { return UNCHOKE_MSG_ID }
func (m *Interested) Id() int    { return INTERESTED_MSG_ID }
func (m *NotInterested) Id() int { return NOT_INTERESTED_MSG_ID }
func (m *Have) Id() int          { return HAVE_MSG_ID }
func (m *Bitfield) Id() int      { return BITFIELD_MSG_ID }
func (m *Request) Id() int       { return REQUEST_MSG_ID }
func (m *Piece) Id() int         { return PIECE_MSG_ID }
func (m *Cancel) Id() int        { return CANCEL_MSG_ID }
func (m *Port) Id() int          { return PORT_MSG_ID }

func (m *Generic) String() string {
	return fmt.Sprintf("Generic{id=%d len(payload)=%d}", m.id, len(m.Payload()))
}

func (m *Choke) String() string {
	return "Choke{}"
}
func (m *Unchoke) String() string {
	return "Unchoke{}"
}
func (m *Interested) String() string {
	return "Interested{}"
}
func (m *NotInterested) String() string {
	return "NotInterested{}"
}

func (m *Have) String() string {
	return fmt.Sprintf("Have{PieceIndex=%d}", m.PieceIndex)
}

func (m *Bitfield) String() string {
	return fmt.Sprintf("Bitfield{len(Bits)=%d}", m.Bits.Length())
}

func (m *Request) String() string {
	return fmt.Sprintf("Request{Index=%d, Begin=%d, Length=%d}",
		m.Index, m.Begin, m.Length)
}

func (m *Piece) String() string {
	return fmt.Sprintf("Piece{Index=%d, Begin=%d, len(Block)=%d}",
		m.Index, m.Begin, len(m.Block))
}

func (m *Cancel) String() string {
	return fmt.Sprintf("Cancel{Index=%d, Begin=%d, Length=%d}",
		m.Index, m.Begin, m.Length)
}
