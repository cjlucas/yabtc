package tracker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPeers(t *testing.T) {
	Convey("When given a peer list in binary format", t, func() {
		// ip is first 4 bytes, port is last 2 bytes (big endian)
		raw := []byte{
			1, 2, 3, 4, 100, 10, // peer 1
			50, 100, 150, 200, 200, 0, // peer 2
		}
		resp := AnnounceResponse{RawPeers: raw}
		Convey("It should parse the list correctly", func() {
			peers := resp.Peers()
			So(len(peers), ShouldEqual, 2)
			So(peers[0].Ip(), ShouldEqual, "1.2.3.4")
			So(peers[0].Port(), ShouldEqual, 25610)
			So(peers[1].Ip(), ShouldEqual, "50.100.150.200")
			So(peers[1].Port(), ShouldEqual, 51200)
		})
	})

	Convey("When given a peer list in dict format", t, func() {
		raw := []byte("ld2:ip7:1.2.3.47:peer id7:peerid14:porti25610eed2:ip14:50.100.150.2007:peer id7:peerid24:porti51200eee")
		resp := AnnounceResponse{RawPeers: raw}
		Convey("It should parse the list correctly", func() {
			peers := resp.Peers()
			So(len(peers), ShouldEqual, 2)
			So(peers[0].Ip(), ShouldEqual, "1.2.3.4")
			So(peers[0].Port(), ShouldEqual, 25610)
			So(peers[1].Ip(), ShouldEqual, "50.100.150.200")
			So(peers[1].Port(), ShouldEqual, 51200)
		})
	})

	Convey("When given a nil peer list", t, func() {
		resp := AnnounceResponse{}
		Convey("It should return a nil slice", func() {
			So(resp.Peers(), ShouldBeNil)
		})
	})
}
