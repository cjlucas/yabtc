package messages

import (
	"encoding/binary"
	"github.com/cjlucas/yabtc/piece"
	"math"
)

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

func EncodeRequestPayload(index uint32, begin uint32, length uint32) []byte {
	var payload [12]byte

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return payload[0:]
}

func DecodeRequestPayload(payload []byte) (index uint32, begin uint32, length uint32) {

	index = binary.BigEndian.Uint32(payload[0:4])
	begin = binary.BigEndian.Uint32(payload[4:8])
	length = binary.BigEndian.Uint32(payload[8:12])

	return index, begin, length
}

func EncodeCancelPayload(index uint32, begin uint32, length uint32) []byte {
	return EncodeRequestPayload(index, begin, length)
}

func DecodeCancelPayload(payload []byte) (index uint32, begin uint32, length uint32) {
	return DecodeRequestPayload(payload)
}

func EncodePiecePayload(index uint32, begin uint32, block []byte) []byte {
	payload := make([]byte, 8+len(block))

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	copy(payload[8:], block)

	return payload
}

func DecodePiecePayload(payload []byte) (index uint32, begin uint32, block []byte) {
	index = binary.BigEndian.Uint32(payload[0:4])
	begin = binary.BigEndian.Uint32(payload[4:8])
	block = payload[8:]

	return index, begin, block
}

func EncodePortPayload(payload uint32) []byte {
	var out [4]byte
	binary.BigEndian.PutUint32(out[0:], payload)
	return out[0:]
}

func DecodePortPayload(payload []byte) uint32 {
	return binary.BigEndian.Uint32(payload)
}
