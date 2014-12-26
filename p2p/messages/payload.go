package messages

import (
	"encoding/binary"
	"github.com/cjlucas/yabtc/piece"
	"math"
)

type RequestPayload struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

type PiecePayload struct {
	Index uint32
	Begin uint32
	Block []byte
}

type CancelPayload RequestPayload

func EncodeHavePayload(payload uint32) []byte {
	var out [4]byte
	binary.BigEndian.PutUint32(out[0:], payload)
	return out[0:]
}

func DecodeHavePayload(payload []byte) uint32 {
	return binary.BigEndian.Uint32(payload)
}

func EncodeBitfieldPayload(pieces []piece.Piece) []byte {
	numPieces := len(pieces)
	bitfieldSize := uint32(math.Ceil(float64(numPieces) / float64(8)))
	b := make([]byte, bitfieldSize)

	for i := 0; i < numPieces; i++ {
		curByte := &b[i/8]
		mod := i % 8
		if pieces[i].Have {
			*curByte |= (byte(1) << (uint32(7 - mod)))
		}
	}

	return b
}

// pieces is expected to be a slice of valid Piece objects
func DecodeBitfieldPayload(payload []byte, pieces []piece.Piece) {
	numPieces := len(pieces)

	for i := 0; i < numPieces; i++ {
		piece := &pieces[i]
		curByte := payload[i/8]
		mod := i % 8
		piece.Have = (curByte >> uint(7-mod) & 0x1) == 1
	}
}

func EncodeRequestPayload(payload *RequestPayload) []byte {
	var buf [12]byte

	binary.BigEndian.PutUint32(buf[0:4], uint32(payload.Index))
	binary.BigEndian.PutUint32(buf[4:8], uint32(payload.Begin))
	binary.BigEndian.PutUint32(buf[8:12], uint32(payload.Length))

	return buf[0:]
}

func DecodeRequestPayload(payload []byte) *RequestPayload {
	var out RequestPayload

	out.Index = binary.BigEndian.Uint32(payload[0:4])
	out.Begin = binary.BigEndian.Uint32(payload[4:8])
	out.Length = binary.BigEndian.Uint32(payload[8:12])

	return &out
}

func EncodeCancelPayload(payload *CancelPayload) []byte {
	var buf [12]byte

	binary.BigEndian.PutUint32(buf[0:4], uint32(payload.Index))
	binary.BigEndian.PutUint32(buf[4:8], uint32(payload.Begin))
	binary.BigEndian.PutUint32(buf[8:12], uint32(payload.Length))

	return buf[0:]
}

func DecodeCancelPayload(payload []byte) *CancelPayload {
	var out CancelPayload

	out.Index = binary.BigEndian.Uint32(payload[0:4])
	out.Begin = binary.BigEndian.Uint32(payload[4:8])
	out.Length = binary.BigEndian.Uint32(payload[8:12])

	return &out
}

func EncodePiecePayload(payload *PiecePayload) []byte {
	buf := make([]byte, 8+len(payload.Block))

	binary.BigEndian.PutUint32(buf[0:4], uint32(payload.Index))
	binary.BigEndian.PutUint32(buf[4:8], uint32(payload.Begin))
	copy(buf[8:], payload.Block)

	return buf
}

func DecodePiecePayload(payload []byte) *PiecePayload {
	var out PiecePayload

	out.Index = binary.BigEndian.Uint32(payload[0:4])
	out.Begin = binary.BigEndian.Uint32(payload[4:8])
	out.Block = payload[8:]

	return &out
}

func EncodePortPayload(payload uint32) []byte {
	var out [4]byte
	binary.BigEndian.PutUint32(out[0:], payload)
	return out[0:]
}

func DecodePortPayload(payload []byte) uint32 {
	return binary.BigEndian.Uint32(payload)
}
