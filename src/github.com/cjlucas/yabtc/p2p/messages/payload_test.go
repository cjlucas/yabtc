package messages

import (
	"testing"

	"github.com/cjlucas/yabtc/bitfield"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChokePayload(t *testing.T) {
	Convey("When given a valid Choke object", t, func() {
		msg := &Choke{}
		Convey("it should produce a valid payload", func() {
			So(len(msg.Payload()), ShouldEqual, 0)
		})
	})
}

func TestUnchokePayload(t *testing.T) {
	Convey("When given a valid Unchoke object", t, func() {
		msg := &Unchoke{}
		Convey("it should produce a valid payload", func() {
			So(len(msg.Payload()), ShouldEqual, 0)
		})
	})
}

func TestInterestedPayload(t *testing.T) {
	Convey("When given a valid Interested object", t, func() {
		msg := &Interested{}
		Convey("it should produce a valid payload", func() {
			So(len(msg.Payload()), ShouldEqual, 0)
		})
	})
}

func TestNotInterestedPayload(t *testing.T) {
	Convey("When given a valid NotInterested object", t, func() {
		msg := &NotInterested{}
		Convey("it should produce a valid payload", func() {
			So(len(msg.Payload()), ShouldEqual, 0)
		})
	})
}

func TestHavePayload(t *testing.T) {
	Convey("When given a valid Have object", t, func() {
		msg := NewHave(5)
		Convey("it should produce a valid payload", func() {
			expected := []byte{0, 0, 0, 5}
			So(msg.Payload(), ShouldResemble, expected)
		})
	})
}

func TestBitfieldPayload(t *testing.T) {
	Convey("When given a valid Bitfield object", t, func() {
		bits := bitfield.New(5)
		for i := 0; i < bits.Length(); i++ {
			bits.Set(i, 1)
		}
		msg := NewBitfield(bits)
		Convey("it should produce a valid payload", func() {
			expected := []byte{0xf8}
			So(msg.Payload(), ShouldResemble, expected)
		})
	})
}

func TestRequestPayload(t *testing.T) {
	Convey("When given a valid Request object", t, func() {
		msg := NewRequest(1, 2, 3)
		Convey("it should produce a valid payload", func() {
			expected := []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
				0, 0, 0, 3}
			So(msg.Payload(), ShouldResemble, expected)
		})
	})
}

func TestPiecePayload(t *testing.T) {
	Convey("When given a valid Piece object", t, func() {
		msg := NewPiece(1, 2, []byte{0xff})
		Convey("it should produce a valid payload", func() {
			expected := []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
				0xff}
			So(msg.Payload(), ShouldResemble, expected)
		})
	})
}

func TestCancelPayload(t *testing.T) {
	Convey("When given a valid Cancel object", t, func() {
		msg := NewCancel(1, 2, 3)
		Convey("it should produce a valid payload", func() {
			expected := []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
				0, 0, 0, 3}
			So(msg.Payload(), ShouldResemble, expected)
		})
	})
}

func TestPortPayload(t *testing.T) {
	Convey("When given a valid Port object", t, func() {
		msg := NewPort(255)
		Convey("it should produce a valid payload", func() {
			expected := []byte{0x00, 0xff}
			So(msg.Payload(), ShouldResemble, expected)
		})
	})
}

/*
 *func TestDecodeHavePayload(t *testing.T) {
 *    Convey("When given a payload [0x0, 0x0, 0x0, 0x1]", t, func() {
 *        payload := []byte{0, 0, 0, 1}
 *
 *        Convey("It should decode to 1", func() {
 *            So(DecodeHavePayload(payload), ShouldEqual, 1)
 *        })
 *    })
 *}
 *
 *func TestEncodeHavePayload(t *testing.T) {
 *    Convey("When given an integer 255", t, func() {
 *        i := 255
 *
 *        Convey("It should encode a valid payload", func() {
 *            expected := []byte{0x0, 0x0, 0x0, 0xFF}
 *            So(EncodeHavePayload(uint32(i)), ShouldResemble, expected)
 *        })
 *    })
 *}
 *
 *func TestEncodeBitfieldPayload(t *testing.T) {
 *    Convey("When given a Piece slice holding 8 members", t, func() {
 *        pieces := make([]piece.Piece, 8)
 *
 *        Convey("and Have == true for all pieces", func() {
 *            for i := 0; i < len(pieces); i++ {
 *                p := &pieces[i]
 *                p.Have = true
 *            }
 *
 *            Convey("It should produce a valid payload", func() {
 *                actual := EncodeBitfieldPayload(pieces)
 *                expected := []byte{0xFF}
 *                So(actual, ShouldResemble, expected)
 *            })
 *        })
 *
 *        Convey("and Have == true for only even numbered pieces", func() {
 *            for i := 0; i < len(pieces); i++ {
 *                p := &pieces[i]
 *                p.Have = i%2 == 0
 *            }
 *
 *            Convey("It should produce a valid payload", func() {
 *                actual := EncodeBitfieldPayload(pieces)
 *                expected := []byte{0xAA} // 10101010
 *                So(actual, ShouldResemble, expected)
 *            })
 *        })
 *    })
 *
 *    Convey("When given a Piece slice holding 9 members", t, func() {
 *        pieces := make([]piece.Piece, 9)
 *
 *        Convey("and Have == true for all pieces", func() {
 *            for i := 0; i < len(pieces); i++ {
 *                p := &pieces[i]
 *                p.Have = true
 *            }
 *
 *            Convey("it should product a valid payload", func() {
 *                actual := EncodeBitfieldPayload(pieces)
 *                expected := []byte{0xFF, 0x80} // 11111111 10000000
 *                So(actual, ShouldResemble, expected)
 *            })
 *        })
 *    })
 *}
 *
 *func TestDecodeBitfieldPayload(t *testing.T) {
 *    Convey("When given an 8 bit bitfield", t, func() {
 *        Convey("and every bit is set to 1", func() {
 *            payload := []byte{0xFF}
 *
 *            Convey("it should set Have == true for all pieces", func() {
 *                pieces := make([]piece.Piece, 8)
 *                DecodeBitfieldPayload(payload, pieces)
 *                for _, p := range pieces {
 *                    So(p.Have, ShouldEqual, true)
 *                }
 *            })
 *        })
 *
 *        Convey("and every even bit is set to 1", func() {
 *            payload := []byte{0xAA} // 10101010
 *
 *            Convey("it should set Have == true for every even piece", func() {
 *                pieces := make([]piece.Piece, 8)
 *                DecodeBitfieldPayload(payload, pieces)
 *                for i := 0; i < len(pieces); i += 2 {
 *                    So(pieces[i].Have, ShouldEqual, true)
 *                }
 *            })
 *        })
 *    })
 *
 *    Convey("When given a 9 bit bitfield", t, func() {
 *        Convey("and every bit is set to 1", func() {
 *            payload := []byte{0xFF, 0x80} // 11111111 10000000
 *
 *            Convey("it should set Have == true for all pieces", func() {
 *                pieces := make([]piece.Piece, 9)
 *                DecodeBitfieldPayload(payload, pieces)
 *                for _, p := range pieces {
 *                    So(p.Have, ShouldEqual, true)
 *                }
 *            })
 *        })
 *    })
 *}
 *
 *func TestEncodeRequestPayload(t *testing.T) {
 *    Convey("When given a valid input", t, func() {
 *        expectedIndex := uint32(5)
 *        expectedBegin := uint32(4)
 *        expectedLength := uint32(16384)
 *
 *        Convey("it should produce a valid payload", func() {
 *            actual := EncodeRequestPayload(
 *                expectedIndex,
 *                expectedBegin,
 *                expectedLength)
 *            expected := []byte{
 *                0x0, 0x0, 0x0, 0x5, // index
 *                0x0, 0x0, 0x0, 0x4, // begin
 *                0x0, 0x0, 0x40, 0x0, // length
 *            }
 *            So(actual, ShouldResemble, expected)
 *        })
 *    })
 *}
 */
