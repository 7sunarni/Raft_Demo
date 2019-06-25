package raft

import (
	"fmt"
	"strconv"
	"testing"
)

func TestPersistance(t *testing.T) {
	log := StableLog{
		FileName: "F:\\_personal_github\\Raft_Demo\\src\\test.log",
	}
	//m := Entry{
	//	Term:  0,
	//	Index: 1,
	//}
	//log.Write(m)
	fmt.Println(log.FirstIndex())
	fmt.Println(log.LastIndex())
}

func TestLog(t *testing.T) {
	stableLog := StableLog{FileName: "F:\\_personal_github\\Raft_Demo\\src\\test.log"}
	raftLog := RaftLog{
		Stable: stableLog,
	}
	raftLog.AppendEntry(0, 0,
		Entry{
			Term:  0,
			Index: 0,
			Data:  []byte("000"),
		})

	for i := 1; i < 10; i++ {
		raftLog.AppendEntry(int64(i), 0, Entry{
			Term:  0,
			Index: int64(i),
			Data:  []byte(strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i)),
		})
	}
	raftLog.NewTermAppendEntry(10, 1,
		Entry{
			Term:  1,
			Index: 10,
			Data:  []byte("AAA"),
		})

	for i := 11; i < 20; i++ {
		raftLog.AppendEntry(int64(i), 1, Entry{
			Term:  1,
			Index: int64(i),
			Data:  []byte(strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i)),
		})
	}

	e1 := Entry{
		Term:  1,
		Index: 20,
		Data:  []byte("20"),
	}
	e2 := Entry{
		Term:  2,
		Index: 21,
		Data:  []byte("21"),
	}
	isOk, i, t2 := raftLog.AppendEntry(20, 1, e1, e2)
	t.Log(isOk, i, t2)
	raftLog.AppendEntry(22, 2, Entry{
		Term:  2,
		Index: 22,
		Data:  []byte("22"),
	})
	t.Log(raftLog)
	raftLog.AppendSnapshot(12, 0, Snapshot{
		Index: 12,
	})
	t.Logf("raftLog.Committed %v \n", raftLog.Committed)
}
