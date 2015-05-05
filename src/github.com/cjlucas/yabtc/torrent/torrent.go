package torrent

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"

	"github.com/zeebo/bencode"
)

type Piece struct {
	Index      int
	ByteOffset int
	Length     int
	Hash       []byte
}

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
	RawInfo      bencode.RawMessage `bencode:"info"`
	Info         Info
	Announce     string `bencode:"announce"`
	CreationDate int64  `bencode:"creation date"`
	Comment      string `bencode:"comment"`
	CreatedBy    string `bencode:"created by"`
	Encoding     string `bencode:"encoding"`
	Pieces       []Piece
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

	if err := bencode.DecodeBytes(m.RawInfo, &m.Info); err != nil {
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
	sha := sha1.New()
	encoder := bencode.NewEncoder(sha)

	if err := encoder.Encode(&m.RawInfo); err != nil {
		panic(err)
	}

	return sha.Sum(nil)
}

func (m *MetaData) InfoHashString() string {
	return fmt.Sprintf("%02X", m.InfoHash())
}

// TODO: make function private
// User should grab Pieces field directly
func (m *MetaData) GeneratePieces() []Piece {
	if m.Pieces != nil {
		return m.Pieces
	}

	numPieces := m.NumPieces()
	m.Pieces = make([]Piece, numPieces)

	files := m.Files()

	curByteOffset := 0
	for i := 0; i < numPieces; i++ {
		p := &m.Pieces[i]
		p.Hash = make([]byte, sha1.Size)

		p.Index = i
		if isLastPiece := i == numPieces-1; isLastPiece {
			p.Length = files.TotalLength() % m.PieceSize()
		} else {
			p.Length = m.PieceSize()
		}
		p.ByteOffset = curByteOffset
		copy(p.Hash, m.Info.Pieces[i*20:(i+1)*20])

		curByteOffset += p.Length
	}

	return m.Pieces
}
