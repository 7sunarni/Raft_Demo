package raft

import (
	"sync"
	"time"
)

const (
	Info  = iota
	Warn  = iota
	Error = iota
)

// TODO  在里面用fmt来打印对象
func NewRaftDebugLog() *RaftDebugLog {
	// TODO 加锁
	return &RaftDebugLog{
		LogMutex: sync.Mutex{},
		Items:    []LogItem{},
	}
}

type RaftDebugLog struct {
	LogMutex sync.Mutex
	Items    []LogItem
}
type LogItem struct {
	Type      int64
	Value     string
	TimeStamp int64
}

func (r *RaftDebugLog) Info(s string) {
	item := LogItem{
		Type:      Info,
		TimeStamp: time.Now().Unix(),
		Value:     s,
	}
	r.Items = append(r.Items, item)

}

func (r *RaftDebugLog) Warn(s string) {
	item := LogItem{
		Type:      Warn,
		TimeStamp: time.Now().Unix(),
		Value:     s,
	}
	r.Items = append(r.Items, item)
}

func (r *RaftDebugLog) Error(s string) {
	item := LogItem{
		Type:      Error,
		TimeStamp: time.Now().Unix(),
		Value:     s,
	}
	r.Items = append(r.Items, item)
}

// 把日志读取到出来返回给前端展示
func (r *RaftDebugLog) GetLogs() []LogItem {
	items := r.Items
	r.Items = []LogItem{}
	return items
}
