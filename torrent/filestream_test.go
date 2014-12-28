package torrent

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var goPath = os.Getenv("GOPATH")

var norm = []File{
	File{[]string{"file1.mp3"}, 1000, ""},
	File{[]string{"file2.mp3"}, 500, ""},
	File{[]string{"file3.mp3"}, 200, ""},
}

var simpleFileStream = FileStream{goPath, norm}

func TestNewFileStream(t *testing.T) {
	Convey("When given valid arguments", t, func() {
		Convey("It should return a valid FileStream", func() {
			actual := NewFileStream(goPath, norm)
			So(actual.Root, ShouldEqual, goPath)
			So(actual.Files, ShouldResemble, norm)
		})
	})
}

func TestTotalLength(t *testing.T) {
	Convey("When given valid arguments", t, func() {
		in := simpleFileStream
		Convey("It should return the sum of each file's length", func() {
			So(in.TotalLength(), ShouldEqual, 1700)
		})
	})
}

func TestBlockValid(t *testing.T) {
	fs := simpleFileStream
	Convey("When given a Block with a negative offset", t, func() {
		block := Block{-1, 500}

		Convey("It should be declared invalid", func() {
			So(fs.BlockValid(block), ShouldBeFalse)
		})
	})

	Convey("When given a Block with a negative length", t, func() {
		block := Block{0, -1}

		Convey("It should be declared invalid", func() {
			So(fs.BlockValid(block), ShouldBeFalse)
		})
	})

	Convey("When given a Block with a length of zero", t, func() {
		block := Block{0, 0}

		Convey("It should be declared invalid", func() {
			So(fs.BlockValid(block), ShouldBeFalse)
		})
	})

	Convey("When given a Block with an offset greater than the entire stream", t, func() {
		block := Block{2000, 50}

		Convey("It should be declared invalid", func() {
			So(fs.BlockValid(block), ShouldBeFalse)
		})
	})

	Convey("When given a Block with a length that overflows the end of the steram", t, func() {
		block := Block{1000, 1000}

		Convey("It should be declared invalid", func() {
			So(fs.BlockValid(block), ShouldBeFalse)
		})
	})

	Convey("When given a valid Block in the middle of the stream", t, func() {
		block := Block{0, 1000}

		Convey("It should be declared valid", func() {
			So(fs.BlockValid(block), ShouldBeTrue)
		})
	})

	Convey("When given a valid Block reaching the end of the stream", t, func() {
		block := Block{1000, 700}

		Convey("It should be declared valid", func() {
			So(fs.BlockValid(block), ShouldBeTrue)
		})
	})
}
