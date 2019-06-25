package raft

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestNode(t *testing.T) {
	for i:=0;i<10;i++{
		fmt.Println(rand.Int63n(5))
	}
}
