package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
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

func (fs *FileStream) seekToOffset(blockOffset int) (f *File, fileOffset int) {
	bytesLeft := blockOffset

	for i := range fs.Files {
		fileSize := int(fs.Files[i].Length)
		if bytesLeft < fileSize {
			return &fs.Files[i], bytesLeft
		}

		bytesLeft -= fileSize
	}

	panic(fmt.Errorf("Could not find file offset for byte offset: %d", blockOffset))
}

func (fs *FileStream) WriteBlock(block Block) error {
	return errors.New("WriteBlock() not implemented")
}

func (fs *FileStream) ReadBlock(block Block) ([]byte, error) {
	return nil, errors.New("ReadBlock() not implemented")
}

func (fs *FileStream) CalculatePieceChecksum(piece FullPiece) bool {
	block := Block{piece.ByteOffset, piece.Length}

	data, err := fs.ReadBlock(block)

	if err != nil {
		return false
	}

	return bytes.Equal(piece.Hash, sha1.New().Sum(data))
}
