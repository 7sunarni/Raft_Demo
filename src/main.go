package main

import (
	"os"
	"raft"
)

func main() {

	if len(os.Args) != 1 {
		node := raft.NewNode(os.Args[1])
		go node.Start()
	} else {
		node := raft.NewNode("9001")
		node.SetNodes([]string{"9002", "9003", "9004", "9005"})
		node.SetState(raft.StateLeader)
		go node.Start()
	}
	for {
		continue
	}
}
