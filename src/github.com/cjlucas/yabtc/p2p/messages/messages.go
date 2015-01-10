package messages

import (
	"bytes"
	"encoding/binary"
	"github.com/cjlucas/yabtc/piece"
)

const CHOKE_MSG_ID = 0
const UNCHOKE_MSG_ID = 1
const INTERESTED_MSG_ID = 2
const NOT_INTERESTED_MSG_ID = 3
const HAVE_MSG_ID = 4
const BITFIELD_MSG_ID = 5
const REQUEST_MSG_ID = 6
const PIECE_MSG_ID = 7
const CANCEL_MSG_ID = 8
const PORT_MSG_ID = 9

type Message struct {
	Len     uint32
	Id      byte
	Payload []byte
}

func New(msgId byte, payload []byte) *Message {
	var msg Message
	msg.Len = uint32(1 + len(payload))
	msg.Id = msgId
	msg.Payload = payload

	return &msg
}

func (m *Message) Bytes() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, m.Len)
	buf.WriteByte(m.Id)
	buf.Write(m.Payload)

	return buf.Bytes()
}

func ChokeMessage() *Message {
	return New(CHOKE_MSG_ID, nil)
}

func UnchokeMessage() *Message {
	return New(UNCHOKE_MSG_ID, nil)
}

func InterestedMessage() *Message {
	return New(INTERESTED_MSG_ID, nil)
}

func NotInterestedMessage() *Message {
	return New(NOT_INTERESTED_MSG_ID, nil)
}

func HaveMessage(pieceIndex uint) *Message {
	return New(HAVE_MSG_ID, EncodeHavePayload(uint32(pieceIndex)))
}

func BitfieldMessage(pieces []piece.Piece) *Message {
	return New(BITFIELD_MSG_ID, EncodeBitfieldPayload(pieces))
}

func RequestMessage(index uint, begin uint, length uint) *Message {
	payload := EncodeRequestPayload(
		uint32(index),
		uint32(begin),
		uint32(length))
	return New(REQUEST_MSG_ID, payload)
}

func PieceMessage(index uint, begin uint, block []byte) *Message {
	payload := EncodePiecePayload(uint32(index), uint32(begin), block)
	return New(PIECE_MSG_ID, payload)
}

func CancelMessage(index uint, begin uint, length uint) *Message {
	payload := EncodeCancelPayload(
		uint32(index),
		uint32(begin),
		uint32(length))
	return New(CANCEL_MSG_ID, payload)
}

func PortMessage(port uint) *Message {
	return New(PORT_MSG_ID, EncodePortPayload(uint32(port)))
}
