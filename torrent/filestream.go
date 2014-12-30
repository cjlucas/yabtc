package torrent

import (
	"crypto/sha1"
	"errors"
	"os"
	"path"
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
	Files FileList
}

type fileAccessPoint struct {
	File          *File
	Offset        int
	BytesExpected int
}

func NewFileStream(root string, files []File) *FileStream {
	return &FileStream{root, files}
}

func (fs *FileStream) BlockValid(block Block) bool {
	return block.Offset >= 0 &&
		block.Length > 0 &&
		block.Offset+block.Length <= fs.Files.TotalLength()
}

func (fs *FileStream) FilePathFromRoot(f *File) string {
	return path.Join(fs.Root, f.Path())
}

func (fs *FileStream) nextFile(curFile *File) *File {
	for i := range fs.Files {
		if curFile == &fs.Files[i] {
			return &fs.Files[i+1]
		}
	}

	return nil
}

func openFileAndSeek(fpath string, seekPos int, mode int) (*os.File, error) {
	fi, err := os.OpenFile(fpath, mode, 0644)

	if err != nil {
		return nil, err
	}

	if _, err := fi.Seek(int64(seekPos), os.SEEK_SET); err != nil {
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

func (fs *FileStream) WriteBlock(block Block, data []byte) error {
	if !fs.BlockValid(block) {
		panic("Received an invalid block")
	}

	if len(data) != block.Length {
		panic("Length of data does not match block.Length")
	}

	bytesWritten := 0
	for _, p := range fs.determineAccessPoints(block) {
		fpath := fs.FilePathFromRoot(p.File)
		if fp, err := openFileAndSeek(fpath, p.Offset, os.O_WRONLY); err != nil {
			return err
		} else {
			n, err := fp.Write(data[bytesWritten:])
			fp.Close()

			if err != nil {
				return err
			}

			if n != p.BytesExpected {
				return errors.New("Wrote an unexpected amount of data")
			}

			bytesWritten += n
		}
	}

	return nil
}

func (fs *FileStream) ReadBlock(block Block) ([]byte, error) {
	if !fs.BlockValid(block) {
		panic("Received an invalid block")
	}

	data := make([]byte, block.Length)

	bytesRead := 0
	for _, p := range fs.determineAccessPoints(block) {
		fpath := fs.FilePathFromRoot(p.File)
		if fp, err := openFileAndSeek(fpath, p.Offset, os.O_RDONLY); err != nil {
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

func (fs *FileStream) CalculatePieceChecksum(piece FullPiece) []byte {
	block := Block{piece.ByteOffset, piece.Length}

	data, err := fs.ReadBlock(block)

	// If there was any error reading the data, return false
	if err != nil {
		return nil
	}

	sha := sha1.New()
	sha.Write(data)

	return sha.Sum(nil)
}
