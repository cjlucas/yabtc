package torrent

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTotalLength(t *testing.T) {
	Convey("When given a FileList of files", t, func() {
		fl := FileList{
			File{[]string{"file1.mp3"}, 100, ""},
			File{[]string{"file2.mp3"}, 200, ""},
			File{[]string{"file3.mp3"}, 500, ""},
		}

		Convey("It should return the sum of each file's length", func() {
			So(fl.TotalLength(), ShouldEqual, 800)
		})
	})
}

func TestPath(t *testing.T) {
	Convey("When given a file located at the root", t, func() {
		f := File{[]string{"file1.mp3"}, 100, ""}

		Convey("It should return the correct relative path", func() {
			So(f.Path(), ShouldEqual, "file1.mp3")
		})
	})

	Convey("When given a file located within multiple subdirectories", t, func() {
		f := File{[]string{"path", "to", "file1.mp3"}, 100, ""}

		Convey("It should return the correct relative path", func() {
			So(f.Path(), ShouldEqual, "path/to/file1.mp3")
		})
	})
}
