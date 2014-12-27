package messages

import (
	"github.com/cjlucas/yabtc/piece"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDecodeHavePayload(t *testing.T) {
	Convey("When given a payload [0x0, 0x0, 0x0, 0x1]", t, func() {
		payload := []byte{0, 0, 0, 1}

		Convey("It should decode to 1", func() {
			So(DecodeHavePayload(payload), ShouldEqual, 1)
		})
	})
}

func TestEncodeHavePayload(t *testing.T) {
	Convey("When given an integer 255", t, func() {
		i := 255

		Convey("It should encode a valid payload", func() {
			expected := []byte{0x0, 0x0, 0x0, 0xFF}
			So(EncodeHavePayload(uint32(i)), ShouldResemble, expected)
		})
	})
}

func TestEncodeBitfieldPayload(t *testing.T) {
	Convey("When given a Piece slice holding 8 members", t, func() {
		pieces := make([]piece.Piece, 8)

		Convey("and Have == true for all pieces", func() {
			for i := 0; i < len(pieces); i++ {
				p := &pieces[i]
				p.Have = true
			}

			Convey("It should produce a valid payload", func() {
				actual := EncodeBitfieldPayload(pieces)
				expected := []byte{0xFF}
				So(actual, ShouldResemble, expected)
			})
		})

		Convey("and Have == true for only even numbered pieces", func() {
			for i := 0; i < len(pieces); i++ {
				p := &pieces[i]
				p.Have = i%2 == 0
			}

			Convey("It should produce a valid payload", func() {
				actual := EncodeBitfieldPayload(pieces)
				expected := []byte{0xAA} // 10101010
				So(actual, ShouldResemble, expected)
			})
		})
	})

	Convey("When given a Piece slice holding 9 members", t, func() {
		pieces := make([]piece.Piece, 9)

		Convey("and Have == true for all pieces", func() {
			for i := 0; i < len(pieces); i++ {
				p := &pieces[i]
				p.Have = true
			}

			Convey("it should product a valid payload", func() {
				actual := EncodeBitfieldPayload(pieces)
				expected := []byte{0xFF, 0x80} // 11111111 10000000
				So(actual, ShouldResemble, expected)
			})
		})
	})
}

func TestDecodeBitfieldPayload(t *testing.T) {
	Convey("When given an 8 bit bitfield", t, func() {
		Convey("and every bit is set to 1", func() {
			payload := []byte{0xFF}

			Convey("it should set Have == true for all pieces", func() {
				pieces := make([]piece.Piece, 8)
				DecodeBitfieldPayload(payload, pieces)
				for _, p := range pieces {
					So(p.Have, ShouldEqual, true)
				}
			})
		})

		Convey("and every even bit is set to 1", func() {
			payload := []byte{0xAA} // 10101010

			Convey("it should set Have == true for every even piece", func() {
				pieces := make([]piece.Piece, 8)
				DecodeBitfieldPayload(payload, pieces)
				for i := 0; i < len(pieces); i += 2 {
					So(pieces[i].Have, ShouldEqual, true)
				}
			})
		})
	})

	Convey("When given a 9 bit bitfield", t, func() {
		Convey("and every bit is set to 1", func() {
			payload := []byte{0xFF, 0x80} // 11111111 10000000

			Convey("it should set Have == true for all pieces", func() {
				pieces := make([]piece.Piece, 9)
				DecodeBitfieldPayload(payload, pieces)
				for _, p := range pieces {
					So(p.Have, ShouldEqual, true)
				}
			})
		})
	})
}

func TestEncodeRequestPayload(t *testing.T) {
	Convey("When given a valid input", t, func() {
		expectedIndex := uint32(5)
		expectedBegin := uint32(4)
		expectedLength := uint32(16384)

		Convey("it should produce a valid payload", func() {
			actual := EncodeRequestPayload(
				expectedIndex,
				expectedBegin,
				expectedLength)
			expected := []byte{
				0x0, 0x0, 0x0, 0x5,
				0x0, 0x0, 0x0, 0x4,
				0x0, 0x0, 0x40, 0x0,
			}
			So(actual, ShouldResemble, expected)
		})
	})
}
