package torrent

import "path"

type File struct {
	PathComponents []string `bencode:"path"`
	Length         int      `bencode:"length"`
	MD5sum         string   `bencode:"md5sum"`
}

type FileList []File

func (f *File) Path() string {
	return path.Join(f.PathComponents...)
}

func (f *File) PathFromRoot(root string) string {
	return path.Join(root, f.Path())
}

func (fl *FileList) TotalLength() int {
	total := 0
	for _, f := range *fl {
		total += int(f.Length)
	}

	return total
}
