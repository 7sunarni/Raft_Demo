package raft

import (
	"testing"
)

const (
	File = "F:\\_personal_github\\Raft_Demo\\src\\test.log"
)

/*
实现的测试用例：
1. 直接向unstable中 append entry
	1.1 entry的term和之前的一样
	1.2 entry的term比之前的大一个
	1.3 entry的index比之前的小，
	1.4 entry的index和之前相同
	1.5 entry的index比之前的大
	1.6 entry的index比之前的大很多

2. 向stable中添加数据，然后向unstable中添加数据
	2.1 添加后向unstable中写入正常的数据
	2.2 添加后向unstable中写入异常的数据

*/

func NewLog() *RaftLog {
	stable := StableLog{
		FileName: File,
	}
	raftLog := RaftLog{
		Stable:   stable,
		Unstable: UnstableLog{},
	}
	return &raftLog
}

// 测试初始化状态
func TestInit(t *testing.T) {
	log := NewLog()
	term, index := log.LastIndexAndTerm()
	if term != -1 || index != -1 {
		t.Fail()
	}
}

func TestAppend(t *testing.T) {
	log := NewLog()

	// 第一次插入数据
	{
		term, index := log.LastIndexAndTerm()
		entry := Entry{
			Term:  term,
			Index: index + 1,
			Data:  []byte{byte(index)},
		}
		log.AppendEntry(entry.Term, entry.Index, entry)
		checkTerm, checkIndex := log.LastIndexAndTerm()
		if checkTerm != term || checkIndex != index+1 {
			t.Fatal("first insert error", checkTerm, checkIndex)
		}
	}
	t.Log("insert first")

	// 持续插入数据
	{
		term, index := log.LastIndexAndTerm()
		for i := index + 1; i < index+10; i++ {
			entry := Entry{
				Term:  term,
				Index: int64(i),
				Data:  []byte{byte(i)},
			}
			log.AppendEntry(entry.Term, entry.Index, entry)
			checkTerm, checkIndex := log.LastIndexAndTerm()
			if term != term || checkIndex != int64(i) {
				t.Fatal("continuous 1st insert error", checkTerm, checkIndex)
			}
		}
		t.Log("continuous 1st insert success")
	}

	// 持续第二次插入数据
	{
		term, index := log.LastIndexAndTerm()
		for i := index + 1; i < index+10; i++ {
			entry := Entry{
				Term:  term + 1,
				Index: int64(i),
				Data:  []byte{byte(i)},
			}
			log.AppendEntry(entry.Term, entry.Index, entry)
			checkTerm, checkIndex := log.LastIndexAndTerm()
			if term != checkTerm-1 || checkIndex != int64(i) {
				t.Fatal("continuous 2nd insert error", checkTerm, checkIndex)
			}
		}
		t.Log("continuous 2nd insert success")
	}
}

// 测试向持久化文件中写入日志
func TestStableLog(t *testing.T) {
	log := NewLog()
	//StableWriteOne()
	//{
	//	term, index := log.LastIndexAndTerm()
	//	if term != 0 || index != 0 {
	//		t.Fatal("stable log error", term, index)
	//	}
	//}
	//StableWriteList()
	{
		term, index := log.LastIndexAndTerm()
		if term != 0 || index != 9 {
			t.Fatal("stable log error", term, index)
		}
	}
	t.Log("stable log success")
}

func StableWriteOne() {
	log := NewLog()
	log.Stable.Write(Entry{
		Term:  0,
		Index: 0,
		Data:  []byte{1},
	})
}

func StableWriteList() {
	log := NewLog()
	for i := 1; i < 10; i++ {
		log.Stable.Write(Entry{
			Term:  int64(0),
			Index: int64(i),
			Data:  []byte{byte(i)},
		})
	}
}

func TestWithStable(t *testing.T) {
	log := NewLog()
	// 测试添加失败的情况
	{
		entry := Entry{
			Term:  0,
			Index: 0,
		}
		log.AppendEntry(entry.Term, entry.Index, entry)
		term, index := log.LastIndexAndTerm()
		if term != 0 && index != 9 {
			t.Fatal("append error", term, index)
		}
	}

	// 测试添加成功的情况
	{
		entry := Entry{
			Term:  2,
			Index: 10,
		}
		log.AppendEntry(entry.Term, entry.Index, entry)
		term, index := log.LastIndexAndTerm()
		if term != 0 && index != 10 {
			t.Fatal("append error", term, index)
		}
	}

}

func TestMoreIndex(t *testing.T) {
	log := NewLog()
	term, index := log.LastIndexAndTerm()
	entry := Entry{
		Term:  term,
		Index: index + 10,
	}
	log.AppendEntry(entry.Term, entry.Index, entry)

	checkTerm, checkIndex := log.LastIndexAndTerm()
	if checkIndex != index {
		t.Fatal("cross append", checkTerm, index)
	}

}
