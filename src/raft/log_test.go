package raft

import (
	"fmt"
	"strconv"
	"testing"
)

func TestPersist(t *testing.T) {
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

// 测试添加第一笔数据，添加第一笔数据的term为-1，index为-1
func TestLogFistInsert(t *testing.T) {
	stableLog := StableLog{FileName: "F:\\_personal_github\\Raft_Demo\\src\\test.log"}
	raftLog := RaftLog{
		Stable: stableLog,
	}
	isOk, i, i2 := raftLog.AppendEntry(-1, -1,
		Entry{
			Term:  0,
			Index: 0,
			Data:  []byte("000"),
		})
	t.Log(isOk, i, i2)
}

// 测试持续添加数据
func TestContinuousInsert(t *testing.T) {
	stableLog := StableLog{FileName: "F:\\_personal_github\\Raft_Demo\\src\\test.log"}
	raftLog := RaftLog{
		Stable: stableLog,
	}
	raftLog.AppendEntry(-1, -1,
		Entry{
			Term:  0,
			Index: 0,
			Data:  []byte("000"),
		})
	for i := 1; i < 10; i++ {
		raftLog.AppendEntry(0, int64(i-1), Entry{
			Term:  0,
			Index: int64(i),
			Data:  []byte(strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i)),
		})
	}
	t.Log(raftLog)
}

// 测试添加数据跨越数据
func TestAddEntryCrossTerm(t *testing.T) {
	stableLog := StableLog{FileName: "F:\\_personal_github\\Raft_Demo\\src\\test.log"}
	raftLog := RaftLog{
		Stable: stableLog,
	}
	raftLog.AppendEntry(-1, -1,
		Entry{
			Term:  0,
			Index: 0,
			Data:  []byte("000"),
		})
	for i := 1; i < 10; i++ {
		raftLog.AppendEntry(0, int64(i-1), Entry{
			Term:  0,
			Index: int64(i),
			Data:  []byte(strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i)),
		})
	}
	raftLog.AppendEntry(0, 9, Entry{
		Term:  1,
		Index: 10,
		Data:  []byte("101010"),
	})
	for i := 11; i < 20; i++ {
		raftLog.AppendEntry(1, int64(i-1), Entry{
			Term:  1,
			Index: int64(i),
			Data:  []byte(strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i)),
		})
	}
	t.Log(raftLog)
	raftLog.Unstable.ShrinkEntry(14)
	t.Log(raftLog)
}
