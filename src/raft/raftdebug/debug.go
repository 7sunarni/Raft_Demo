package raftdebug

import (
	"fmt"
	"sync"
	"time"
)

const (
	Trace = iota
	Info
	Warn
	Error
	Fatal
)

func NewRaftDebugLog() *RaftDebugLog {
	return &RaftDebugLog{
		m:     sync.Mutex{},
		items: []LogItem{},
	}
}

type RaftDebugLog struct {
	m     sync.Mutex
	items []LogItem
}
type LogItem struct {
	Type      int64
	Value     string
	TimeStamp int64
}

func (r *RaftDebugLog) Trace(a ...interface{}) {
	item := LogItem{
		Type:      Trace,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.m.Lock()
	defer r.m.Unlock()
	r.items = append(r.items, item)

}

func (r *RaftDebugLog) Info(a ...interface{}) {
	item := LogItem{
		Type:      Info,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.m.Lock()
	defer r.m.Unlock()
	r.items = append(r.items, item)

}

func (r *RaftDebugLog) Warn(a ...interface{}) {
	item := LogItem{
		Type:      Warn,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.m.Lock()
	defer r.m.Unlock()
	r.items = append(r.items, item)
}

func (r *RaftDebugLog) Error(a ...interface{}) {
	item := LogItem{
		Type:      Error,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.m.Lock()
	defer r.m.Unlock()
	r.items = append(r.items, item)
}

func (r *RaftDebugLog) Fatal(a ...interface{}) {
	item := LogItem{
		Type:      Fatal,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.m.Lock()
	defer r.m.Unlock()
	r.items = append(r.items, item)
}

// 把日志读取到出来返回给前端展示
func (r *RaftDebugLog) GetLogs() []LogItem {
	r.m.Lock()
	defer r.m.Unlock()
	items := r.items
	r.items = []LogItem{}
	return items
}
