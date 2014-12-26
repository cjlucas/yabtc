package torrent

import (
	"path"
)

type File struct {
	PathComponents []string `bencode:"path"`
	Length         int64    `bencode:"length"`
	MD5sum         string   `bencode:"md5sum"`
}

func (f *File) Path() string {
	return path.Join(f.PathComponents...)
}
