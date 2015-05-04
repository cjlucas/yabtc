package messages

import (
	"encoding/binary"
	"errors"

	"github.com/cjlucas/yabtc/bitfield"
)

var invalidPayloadError = errors.New("invalid payload")

func (m *Generic) Payload() []byte       { return m.payload }
func (m *Choke) Payload() []byte         { return nil }
func (m *Unchoke) Payload() []byte       { return nil }
func (m *Interested) Payload() []byte    { return nil }
func (m *NotInterested) Payload() []byte { return nil }

func (m *Have) Payload() []byte {
	var out [4]byte
	binary.BigEndian.PutUint32(out[0:], uint32(m.PieceIndex))
	return out[0:]
}

func (m *Bitfield) Payload() []byte {
	return m.Bits.Bytes()
}

func (m *Request) Payload() []byte {
	var payload [12]byte

	binary.BigEndian.PutUint32(payload[0:4], uint32(m.Index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(m.Begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(m.Length))

	return payload[:]
}

func (m *Cancel) Payload() []byte {
	var payload [12]byte

	binary.BigEndian.PutUint32(payload[0:4], uint32(m.Index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(m.Begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(m.Length))

	return payload[:]
}

func (m *Piece) Payload() []byte {
	payload := make([]byte, 8+len(m.Block))

	binary.BigEndian.PutUint32(payload[0:4], uint32(m.Index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(m.Begin))
	copy(payload[8:], m.Block)

	return payload
}

func (m *Port) Payload() []byte {
	var out [2]byte
	binary.BigEndian.PutUint16(out[:], uint16(m.Port))
	return out[:]
}

func (m *Have) decodePayload(payload []byte) error {
	if len(payload) < 4 {
		return invalidPayloadError
	}
	m.PieceIndex = int(binary.BigEndian.Uint32(payload))
	return nil
}

func (m *Bitfield) decodePayload(payload []byte) error {
	m.Bits = bitfield.New(len(payload) * 8)
	m.Bits.SetBytes(payload)
	return nil
}

func (m *Request) decodePayload(payload []byte) error {
	if len(payload) < 12 {
		return invalidPayloadError
	}
	m.Index = int(binary.BigEndian.Uint32(payload[0:4]))
	m.Begin = int(binary.BigEndian.Uint32(payload[4:8]))
	m.Length = int(binary.BigEndian.Uint32(payload[8:12]))
	return nil
}

func (m *Cancel) decodePayload(payload []byte) error {
	if len(payload) < 12 {
		return invalidPayloadError
	}
	m.Index = int(binary.BigEndian.Uint32(payload[0:4]))
	m.Begin = int(binary.BigEndian.Uint32(payload[4:8]))
	m.Length = int(binary.BigEndian.Uint32(payload[8:12]))
	return nil
}

func (m *Piece) decodePayload(payload []byte) error {
	if len(payload) < 8 {
		return invalidPayloadError
	}

	m.Index = int(binary.BigEndian.Uint32(payload[0:4]))
	m.Begin = int(binary.BigEndian.Uint32(payload[4:8]))
	m.Block = payload[8:]

	return nil
}

func (m *Port) decodePayload(payload []byte) error {
	if len(payload) < 4 {
		return invalidPayloadError
	}
	m.Port = int(binary.BigEndian.Uint32(payload))
	return nil
}
