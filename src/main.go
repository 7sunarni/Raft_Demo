package main

import (
	"awesomeProject1/raft"
	"fmt"
	"os"
	"strconv"
)

func main1() {

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

func main() {
	raftLog := raft.RaftLog{}

	raftLog.AppendEntry(0, 0,
		raft.Entry{
			Term:  0,
			Index: 0,
			Data:  []byte("000"),
		})

	for i := 1; i < 10; i++ {
		raftLog.AppendEntry(int64(i), 0, raft.Entry{
			Term:  0,
			Index: int64(i),
			Data:  []byte(strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i)),
		})
	}
	raftLog.NewTermAppendEntry(10, 1,
		raft.Entry{
			Term:  1,
			Index: 10,
			Data:  []byte("AAA"),
		})

	for i := 11; i < 20; i++ {
		raftLog.AppendEntry(int64(i), 1, raft.Entry{
			Term:  1,
			Index: int64(i),
			Data:  []byte(strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i)),
		})
	}

	fmt.Println(raftLog)
	raftLog.AppendSnapshot(7, 0, raft.Snapshot{
		Index:7,
	})
	fmt.Println(raftLog)
}
