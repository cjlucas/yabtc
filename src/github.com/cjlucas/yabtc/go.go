package main

import "fmt"

func main() {
	c := make(chan bool)
	fmt.Println("init goroutines")
	for i := 0; i < 100000; i++ {
		go func(i int, c chan bool) {
			for {
				<-c
			}
		}(i, c)
	}
	fmt.Println("done init")

	for {
		select {
		default:
			c <- true
		}
	}
}
