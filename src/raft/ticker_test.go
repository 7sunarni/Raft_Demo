package raft

import (
	"fmt"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {

	tkr := time.NewTicker(5 * time.Second)

	mChan := make(chan int)

	go passChan(mChan)

	for {
		select {
		case <-tkr.C:
			fmt.Println("tkr active")
		case i := <-mChan:
			fmt.Println("get mChan", i)
			if i == 5 {
				tkr.Stop()
				fmt.Println("tkr close", i)
			}
		}
	}
}

func passChan(c chan int) {
	var i int
	for {
		time.Sleep(3 * time.Second)
		c <- i
		i++
	}
}
