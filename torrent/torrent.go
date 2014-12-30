package torrent

import (
	"crypto/sha1"
	"errors"
	"fmt"
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

type MetaInfo struct {
	Info         Info   `bencode:"info"`
	Announce     string `bencode:"announce"`
	CreationDate int64  `bencode:"creation date"`
	Comment      string `bencode:"comment"`
	CreatedBy    string `bencode:"created by"`
	Encoding     string `bencode:"encoding"`
}

type Torrent struct {
	MetaInfo MetaInfo
	Pieces   []FullPiece
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

func (t *Torrent) NumPieces() int {
	return len(t.MetaInfo.Info.Pieces) / sha1.Size
}

func (t *Torrent) IsMultiFile() bool {
	return len(t.MetaInfo.Info.Files) > 0
}

func (t *Torrent) Files() FileList {
	var files FileList
	info := t.MetaInfo.Info
	if t.IsMultiFile() {
		for _, f := range info.Files {
			newPathComponents := []string{t.MetaInfo.Info.Name}
			newPathComponents = append(newPathComponents, f.PathComponents...)
			f.PathComponents = newPathComponents
			files = append(files, f)
		}
	} else {
		f := File{[]string{info.Name},
			info.Length,
			info.MD5sum}
		files = append(files, f)
	}

	return files
}

func (t *Torrent) InfoHash() []byte {
	info := make(map[string]interface{})

	m := t.MetaInfo

	info["name"] = m.Info.Name
	info["piece length"] = m.Info.PieceLength
	info["pieces"] = m.Info.Pieces
	info["length"] = m.Info.Length

	sha := sha1.New()
	encoder := bencode.NewEncoder(sha)

	if err := encoder.Encode(&info); err != nil {
		panic(err)
	}

	return sha.Sum(nil)
}

func (t *Torrent) InfoHashString() string {
	return fmt.Sprintf("%02X", t.InfoHash())
}

func (t *Torrent) generatePieces() {
	numPieces := len(t.MetaInfo.Info.Pieces) / sha1.Size

	t.Pieces = make([]FullPiece, numPieces)

	curByteOffset := 0
	for i := 0; i < t.NumPieces(); i++ {
		isLastPiece := i == t.NumPieces()-1
		p := &t.Pieces[i]
		p.Hash = make([]byte, sha1.Size)

		p.Index = i
		p.Have = false
		if isLastPiece {
			p.Length = 1162936320 % t.MetaInfo.Info.PieceLength
		} else {
			p.Length = t.MetaInfo.Info.PieceLength
		}
		p.ByteOffset = curByteOffset
		copy(p.Hash, t.MetaInfo.Info.Pieces[i*20:(i+1)*20])

		curByteOffset += p.Length
	}
}
