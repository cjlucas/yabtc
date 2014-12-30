package torrent

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var goPath = os.Getenv("GOPATH")

var norm = FileList{
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

func TestDetermineAccessPoints(t *testing.T) {
	fs := simpleFileStream

	Convey("When given a valid block spanning one file", t, func() {
		block := Block{0, 100}

		Convey("it should return a single correct access point", func() {
			points := fs.determineAccessPoints(block)
			So(len(points), ShouldEqual, 1)

			p := points[0]
			So(*p.File, ShouldResemble, fs.Files[0])
			So(p.Offset, ShouldEqual, 0)
			So(p.BytesExpected, ShouldEqual, 100)
		})
	})

	Convey("When given a valid block spanning multiple files", t, func() {
		block := Block{0, 1700}
		points := fs.determineAccessPoints(block)

		Convey("It should return 3 access points", func() {
			So(len(points), ShouldEqual, 3)
		})

		Convey("It should have the correct access points", func() {
			for i := range fs.Files {
				p := points[i]
				So(*p.File, ShouldResemble, fs.Files[i])
				So(p.BytesExpected, ShouldEqual, fs.Files[i].Length)
				So(p.Offset, ShouldEqual, 0)
			}
		})
	})

	Convey("When given a valid block in the middle of a file stream", t, func() {
		block := Block{1200, 400}
		points := fs.determineAccessPoints(block)

		Convey("It should return 2 access points", func() {
			So(len(points), ShouldEqual, 2)
		})

		Convey("It should return the correct access points", func() {
			So(*points[0].File, ShouldResemble, fs.Files[1])
			So(points[0].Offset, ShouldEqual, 200)
			So(points[0].BytesExpected, ShouldEqual, 300)

			So(*points[1].File, ShouldResemble, fs.Files[2])
			So(points[1].Offset, ShouldEqual, 0)
			So(points[1].BytesExpected, ShouldEqual, 100)
		})
	})
}

func TestFilePathFromRoot(t *testing.T) {
	Convey("When given a filestream consisting of one file", t, func() {
		f := FileList{
			File{[]string{"file1.mp3"}, 100, ""},
		}

		fs := FileStream{"/root", f}
		Convey("It should return the correct path", func() {
			actual := fs.FilePathFromRoot(&f[0])
			So(actual, ShouldEqual, "/root/file1.mp3")
		})
	})

	Convey("When given a filestream consisting of multiple files", t, func() {
		f := FileList{
			File{[]string{"file1.mp3"}, 100, ""},
			File{[]string{"path", "to", "file2.mp3"}, 100, ""},
		}

		fs := FileStream{"/root", f}

		Convey("It should return the correct path for each file", func() {
			So(fs.FilePathFromRoot(&f[0]), ShouldEqual, "/root/file1.mp3")
			So(fs.FilePathFromRoot(&f[1]), ShouldEqual, "/root/path/to/file2.mp3")
		})
	})
}
