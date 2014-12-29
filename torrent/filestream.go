package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"os"
)

type Block struct {
	Offset, Length int
}

// TODO better name
type FullPiece struct {
	Index      int
	Have       bool
	ByteOffset int
	Length     int
	Hash       []byte
}

type FileStream struct {
	Root  string
	Files []File
}

type fileAccessPoint struct {
	File          *File
	Offset        int
	BytesExpected int
}

func NewFileStream(root string, files []File) *FileStream {
	return &FileStream{root, files}
}

func (fs *FileStream) TotalLength() int {
	total := 0
	for i := range fs.Files {
		total += int(fs.Files[i].Length)
	}

	return total
}

func (fs *FileStream) BlockValid(block Block) bool {
	return block.Offset >= 0 &&
		block.Length > 0 &&
		block.Offset+block.Length <= fs.TotalLength()
}

func (fs *FileStream) nextFile(curFile *File) *File {
	for i := range fs.Files {
		if curFile == &fs.Files[i] {
			return &fs.Files[i+1]
		}
	}

	return nil
}

func openFileAndSeek(fpath string, seekPos int) (*os.File, error) {
	fi, err := os.Open(fpath)

	if err != nil {
		return nil, err
	}

	if _, err := fi.Seek(int64(seekPos), 0); err != nil {
		fi.Close()
		return nil, err
	}

	return fi, nil
}

// block must be valid
func (fs *FileStream) determineAccessPoints(block Block) []fileAccessPoint {
	var points []fileAccessPoint
	var curFile *File
	var curOffset int

	// Find start of block
	bytesLeftUntilBlockStart := block.Offset

	for i := range fs.Files {
		fileSize := int(fs.Files[i].Length)
		if bytesLeftUntilBlockStart < fileSize {
			curFile = &fs.Files[i]
			curOffset = bytesLeftUntilBlockStart
			break
		}
		bytesLeftUntilBlockStart -= fileSize
	}

	// Generate acccess points

	bytesLeft := block.Length
	for bytesLeft > 0 {
		var p fileAccessPoint

		bytesLeftInFile := curFile.Length - curOffset
		// If
		if bytesLeft > bytesLeftInFile {
			p = fileAccessPoint{curFile, curOffset, bytesLeftInFile}
			curFile = fs.nextFile(curFile)
			curOffset = 0
			bytesLeft -= bytesLeftInFile
		} else {
			p = fileAccessPoint{curFile, curOffset, bytesLeft}
			bytesLeft -= bytesLeft
		}

		points = append(points, p)
	}

	return points
}

func (fs *FileStream) WriteBlock(block Block) error {
	if !fs.BlockValid(block) {
		panic("Received an invalid block")
	}

	return errors.New("WriteBlock() not implemented")
}

func (fs *FileStream) ReadBlock(block Block) ([]byte, error) {
	if !fs.BlockValid(block) {
		panic("Received an invalid block")
	}

	data := make([]byte, block.Length)

	bytesRead := 0
	for _, p := range fs.determineAccessPoints(block) {
		if fp, err := openFileAndSeek(p.File.Path(), p.Offset); err != nil {
			return nil, err
		} else {
			n, err := fp.Read(data[bytesRead:])
			fp.Close()

			if err != nil {
				return nil, err
			}

			if n != p.BytesExpected {
				return nil, errors.New("Received unexpected amount of data")
			}

			bytesRead += n
		}
	}

	return data, nil
}

func (fs *FileStream) CalculatePieceChecksum(piece FullPiece) bool {
	block := Block{piece.ByteOffset, piece.Length}

	data, err := fs.ReadBlock(block)

	// If there was any error reading the data, return false
	if err != nil {
		return false
	}

	return bytes.Equal(piece.Hash, sha1.New().Sum(data))
}