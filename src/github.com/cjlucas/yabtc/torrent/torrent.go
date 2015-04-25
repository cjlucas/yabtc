package torrent

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"

	"github.com/zeebo/bencode"
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

type MetaData struct {
	Info         Info   `bencode:"info"`
	Announce     string `bencode:"announce"`
	CreationDate int64  `bencode:"creation date"`
	Comment      string `bencode:"comment"`
	CreatedBy    string `bencode:"created by"`
	Encoding     string `bencode:"encoding"`
}

func ParseFile(fname string) (*MetaData, error) {
	if buf, err := ioutil.ReadFile(fname); err != nil {
		return nil, err
	} else {
		return ParseBytes(buf)
	}
}

func ParseBytes(b []byte) (*MetaData, error) {
	var m MetaData
	if err := bencode.DecodeBytes(b, &m); err != nil {
		return nil, fmt.Errorf("bencode error: %s", err)
	}

	return &m, nil
}

func (m *MetaData) NumPieces() int {
	return len(m.Info.Pieces) / sha1.Size
}

func (m *MetaData) IsMultiFile() bool {
	return len(m.Info.Files) > 0
}

func (m *MetaData) PieceSize() int {
	return m.Info.PieceLength
}

func (m *MetaData) Files() FileList {
	var files FileList
	info := m.Info
	if m.IsMultiFile() {
		for _, f := range info.Files {
			newPathComponents := []string{info.Name}
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

func (m *MetaData) InfoHash() []byte {
	info := make(map[string]interface{})

	info["name"] = m.Info.Name
	info["piece length"] = m.Info.PieceLength
	info["pieces"] = m.Info.Pieces
	info["length"] = m.Info.Length
	info["private"] = m.Info.Private

	sha := sha1.New()
	encoder := bencode.NewEncoder(sha)

	if err := encoder.Encode(&info); err != nil {
		panic(err)
	}

	return sha.Sum(nil)
}

func (m *MetaData) InfoHashString() string {
	return fmt.Sprintf("%02X", m.InfoHash())
}

func (m *MetaData) GeneratePieces() (pieces []FullPiece) {
	numPieces := len(m.Info.Pieces) / sha1.Size

	pieces = make([]FullPiece, numPieces)

	files := m.Files()

	curByteOffset := 0
	for i := 0; i < m.NumPieces(); i++ {
		isLastPiece := i == m.NumPieces()-1
		p := &pieces[i]
		p.Hash = make([]byte, sha1.Size)

		p.Index = i
		p.Have = false
		if isLastPiece {
			p.Length = files.TotalLength() % m.PieceSize()
		} else {
			p.Length = m.PieceSize()
		}
		p.ByteOffset = curByteOffset
		copy(p.Hash, m.Info.Pieces[i*20:(i+1)*20])

		curByteOffset += p.Length
	}

	return
}
