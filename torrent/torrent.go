package torrent

import (
	"crypto/sha1"
	"errors"
	"github.com/zeebo/bencode"
	"os"
)

type Info struct {
	Name        string `bencode:"name"`
	Length      int64  `bencode:"length"`
	PieceLength int    `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"`
	Private     int    `bencode:"private"`
	Files       []File `bencode:"files"`
	MD5sum      string `bencode:"md5sum"`
}

// TODO use piece.Piece instead
type Piece struct {
	startByte, endByte int64
	startFile, endFile int
}

type TorrentParser struct {
	MetaInfo MetaInfo
	Pieces   []Piece
}

func (t *TorrentParser) generatePieceInfo() {
	t.Pieces = make([]Piece, t.MetaInfo.NumPieces())
	files := t.MetaInfo.Files()
	curFile := 0
	curByte := int64(0)
	for i, _ := range t.Pieces {
		p := &t.Pieces[i]
		bytesRemainingInPiece := int64(t.MetaInfo.Info.PieceLength)
		p.startFile = curFile
		p.startByte = curByte
		for {
			bytesRemainingInFile := files[curFile].Length - curByte
			if bytesRemainingInFile == 0 {
				break
			}

			if bytesRemainingInPiece < bytesRemainingInFile {
				curByte += bytesRemainingInPiece
				bytesRemainingInPiece = 0
				break
			} else {
				bytesRemainingInPiece -= bytesRemainingInFile
				curByte += bytesRemainingInFile
				if curFile < len(files)-1 {
					curFile++
					curByte = 0
				}
			}
		}

		p.endFile = curFile
		p.endByte = curByte
	}
}

func (tp *TorrentParser) GenerateChecksum(p Piece) ([]byte, error) {
	bytesToRead := make([]int64, p.endFile-p.startFile+1)

	files := tp.MetaInfo.Files()
	// Determine the number of bytes to read for each file
	for i := p.startFile; i <= p.endFile; i++ {
		bytesToReadForFile := &bytesToRead[i-p.startFile]

		if p.startFile == p.endFile {
			*bytesToReadForFile = p.endByte - p.startByte
			break
		}

		info, err := os.Stat(files[i].Path())
		if err != nil {
			return nil, err
		}

		fileSize := info.Size()

		switch i {
		case p.startFile:
			*bytesToReadForFile = fileSize - p.startByte
		case p.endFile:
			*bytesToReadForFile = p.endByte
		default:
			*bytesToReadForFile = fileSize
		}

	}

	sha := sha1.New()

	for i := p.startFile; i <= p.endFile; i++ {
		bufSize := bytesToRead[i-p.startFile]
		buf := make([]byte, bufSize)

		fp, err := os.Open(files[i].Path())
		if err != nil {
			return nil, err
		}

		defer fp.Close()

		switch i {
		case p.startFile:
			fp.ReadAt(buf, p.startByte)
		default:
			fp.Read(buf)
		}

		sha.Write(buf)
	}

	return sha.Sum(nil), nil
}

func ParseFile(fname string) (*TorrentParser, error) {
	stat, err := os.Stat(fname)

	if err != nil {
		return nil, errors.New("Could not stat file")
	}

	buf := make([]byte, stat.Size())

	fp, err := os.Open(fname)

	if err != nil {
		return nil, errors.New("Could not open file")
	}

	defer fp.Close()

	n, err := fp.Read(buf)

	if int64(n) != stat.Size() {
		return nil, errors.New("Error reading file")
	}

	return ParseBytes(buf)
}

func ParseBytes(b []byte) (*TorrentParser, error) {
	var p TorrentParser
	if err := bencode.DecodeBytes(b, &p.MetaInfo); err != nil {
		return nil, errors.New("Error decoding torrent")
	}

	p.generatePieceInfo()
	return &p, nil
}
