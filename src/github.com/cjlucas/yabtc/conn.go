package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "media.cjlucas.net:80")

	if err != nil {
		panic(err)
	}

	c := make(chan bool)
	go func() {
		buf := make([]byte, 10)
		conn.Read(buf)
		c <- true
	}()

	select {
	case <-c:
		return
	case <-time.After(5 * time.Second):
		fmt.Println("timeout")
		//conn.Close()
	}

	<-c
}
