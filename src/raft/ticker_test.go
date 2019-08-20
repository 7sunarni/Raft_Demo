package raft

import (
	"fmt"
	"math/rand"
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

func TestTickerNanotime(t *testing.T) {
	tkr := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-tkr.C:
				tkr = time.NewTicker(time.Duration(10+rand.Int63n(2)) * time.Second)
				t.Log(time.Now().Minute(), time.Now().Second())
				t.Logf("ticker %#v \n %v \n", tkr, tkr)
			}
		}
	}()
	time.Sleep(2 * time.Minute)
}

func TestRand(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		duration := time.Duration(HeartBeatTimeout+rand.Int63n(5)) * time.Second
		t.Log(duration)
	}
}
