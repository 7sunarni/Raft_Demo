package raft

import "time"

/*
progress是描述以leader视角下的follower状态的结构体，
leader通过follower的progress状态来做出不同的响应：heartbeat,app,snapshot等
*/
const (
	// 初始化状态，leader不知道follower状态
	ProgressUnknown = iota
	// 快速备份状态，follower落后leader太多
	ProgressSnapshot
	// follower和leader状态一致，每次有更新直接复制给follower
	ProgressReplicate
)

type Progress struct {
	Node map[string]ProgressState
}

type ProgressState struct {
	Active   bool
	ActiveTS time.Time
	Type     int64
	Index    int64
	Term     int64
}
