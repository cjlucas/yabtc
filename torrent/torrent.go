package torrent

import (
	"crypto/sha1"
	"errors"
	"github.com/zeebo/bencode"
	"io/ioutil"
)

type Info struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	PieceLength int    `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"`
	Private     int    `bencode:"private"`
	Files       []File `bencode:"files"`
	MD5sum      string `bencode:"md5sum"`
}

type Torrent struct {
	MetaInfo MetaInfo
	Pieces   []FullPiece
}

func (t *Torrent) generatePieces() {
	numPieces := len(t.Pieces) / sha1.Size

	t.Pieces = make([]FullPiece, numPieces)

	curByteOffset := 0
	for i := range t.Pieces {
		p := &t.Pieces[i]
		p.Hash = make([]byte, 20)

		p.Index = i
		p.Have = false
		p.Length = t.MetaInfo.Info.PieceLength // TODO: this is incorrect for the final piece
		p.ByteOffset = curByteOffset
		copy(p.Hash, t.MetaInfo.Info.Pieces[i*20:(i+1)*20])

		curByteOffset += p.Length
	}
}

func ParseFile(fname string) (*Torrent, error) {
	if buf, err := ioutil.ReadFile(fname); err != nil {
		return nil, err
	} else {
		return ParseBytes(buf)
	}
}

func ParseBytes(b []byte) (*Torrent, error) {
	var p Torrent
	if err := bencode.DecodeBytes(b, &p.MetaInfo); err != nil {
		return nil, errors.New("Error decoding torrent")
	}

	p.generatePieces()
	return &p, nil
}
