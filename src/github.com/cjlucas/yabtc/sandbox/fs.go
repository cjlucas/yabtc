package main

import (
	"fmt"
	"os"

	"github.com/cjlucas/yabtc/torrent"
)

func main() {
	mdata, err := torrent.ParseFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	inFs := torrent.NewFileStream("/Users/chris/Downloads/tds", mdata.Files())
	outFs := torrent.NewFileStream("/Users/chris/Downloads", mdata.Files())

	fmt.Println(inFs)
	fmt.Println(outFs)

	for _, p := range mdata.GeneratePieces() {
		block := torrent.Block{Offset: p.ByteOffset, Length: p.Length}
		buf, err := inFs.ReadBlock(block)
		if err != nil {
			fmt.Println("error on read")
			panic(err)
		}

		err = outFs.WriteBlock(block, buf)
		if err != nil {
			fmt.Println("error on write")
			panic(err)
		}
	}
}
