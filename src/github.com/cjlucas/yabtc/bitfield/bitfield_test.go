package bitfield

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetBytes(t *testing.T) {
	Convey("When given bytes of valid length", t, func() {
		Convey("It should return the proper bit values", func() {
			cases := []struct {
				input  []byte
				values []int
			}{
				{
					[]byte{0xff, 0xff},
					[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
				{
					[]byte{0x00, 0x00},
					[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				{
					[]byte{0xaa, 0xaa},
					[]int{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0},
				},
			}

			for _, c := range cases {
				b := New(len(c.values))
				b.SetBytes(c.input)
				for i, v := range c.values {
					So(b.Get(i), ShouldEqual, v)
				}
			}
		})
	})
}

func TestSet(t *testing.T) {
	Convey("when setting values at valid indicies", t, func() {
		Convey("It should set the value correctly", func() {
			cases := []struct {
				initial []byte
				vals    []int
			}{
				{
					[]byte{0x00, 0x00},
					[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
				{
					[]byte{0xff, 0xff},
					[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				{
					[]byte{0xaa, 0xaa},
					[]int{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0},
				},
			}

			for _, c := range cases {
				b := New(len(c.vals))
				b.SetBytes(c.initial)
				for i, v := range c.vals {
					b.Set(i, v)
					So(b.Get(i), ShouldEqual, v)
				}
			}
		})
	})
}

func TestBytes(t *testing.T) {
	Convey("When given a valid Bitfield object", t, func() {
		cases := []struct {
			bits     []int
			expected []byte
		}{
			{
				[]int{1, 1, 1, 1, 1, 1, 1, 1},
				[]byte{0xff},
			},
			{
				[]int{0, 0, 0, 0, 0, 0, 0, 0},
				[]byte{0x0},
			},
			{
				[]int{1, 1, 1, 1, 1},
				[]byte{0xf8},
			},
		}
		Convey("It should return a valid byte representation", func() {
			for _, c := range cases {
				bits := New(len(c.bits))
				for i := 0; i < len(c.bits); i++ {
					bits.Set(i, c.bits[i])
				}

				So(bits.Bytes(), ShouldResemble, c.expected)
			}
		})
	})
}
