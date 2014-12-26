package torrent

import (
	"crypto/sha1"
	"fmt"
	"github.com/zeebo/bencode"
)

type MetaInfo struct {
	Info         Info   `bencode:"info"`
	Announce     string `bencode:"announce"`
	CreationDate int64  `bencode:"creation date"`
	Comment      string `bencode:"comment"`
	CreatedBy    string `bencode:"created by"`
	Encoding     string `bencode:"encoding"`
}

func (mp *MetaInfo) IsMultiFile() bool {
	return len(mp.Info.Files) > 0
}

func (mp *MetaInfo) NumPieces() int {
	return len(mp.Info.Pieces) / 20
}

func (mp *MetaInfo) PieceChecksum(piece_num int) []byte {
	return mp.Info.Pieces[piece_num*20 : (piece_num+1)*20]
}

func (mp *MetaInfo) Files() []File {
	m := mp
	var files []File
	info := m.Info
	if m.IsMultiFile() {
		for _, f := range m.Info.Files {
			newPathComponents := []string{m.Info.Name}
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

func (m *MetaInfo) InfoHash() []byte {
	info := make(map[string]interface{})

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

func (m *MetaInfo) InfoHashString() string {
	return fmt.Sprintf("%02X", m.InfoHash())
}
