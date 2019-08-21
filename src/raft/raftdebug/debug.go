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

func (r *RaftDebugLog) Trace(a ...interface{}) {
	item := LogItem{
		Type:      Trace,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.LogMutex.Lock()
	defer r.LogMutex.Unlock()
	r.Items = append(r.Items, item)

}

func (r *RaftDebugLog) Info(a ...interface{}) {
	item := LogItem{
		Type:      Info,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.LogMutex.Lock()
	defer r.LogMutex.Unlock()
	r.Items = append(r.Items, item)

}

func (r *RaftDebugLog) Warn(a ...interface{}) {
	item := LogItem{
		Type:      Warn,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.LogMutex.Lock()
	defer r.LogMutex.Unlock()
	r.Items = append(r.Items, item)
}

func (r *RaftDebugLog) Error(a ...interface{}) {
	item := LogItem{
		Type:      Error,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.LogMutex.Lock()
	defer r.LogMutex.Unlock()
	r.Items = append(r.Items, item)
}

func (r *RaftDebugLog) Fatal(a ...interface{}) {
	item := LogItem{
		Type:      Fatal,
		TimeStamp: time.Now().Unix() * 1000,
		Value:     fmt.Sprint(a),
	}
	r.LogMutex.Lock()
	defer r.LogMutex.Unlock()
	r.Items = append(r.Items, item)
}

// 把日志读取到出来返回给前端展示
func (r *RaftDebugLog) GetLogs() []LogItem {
	r.LogMutex.Lock()
	defer r.LogMutex.Unlock()
	items := r.Items
	r.Items = []LogItem{}
	return items
}
