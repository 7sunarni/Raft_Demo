package raft

import "time"

/*
progress是描述以leader视角下的follower状态的结构体，
leader通过follower的progress状态来做出不同的响应：heartbeat,app,snapshot等
*/
const (
	// 初始化状态，leader不知道follower状态，发出第一条消息后，收到消息回复后
	ProgressUnknown = iota

	/*
	快速备份状态，follower落后leader太多
	处于Snapshot状态的follower，leader先将目前的term，index作为snapshot发给follower
	然后每次leader有新的entry的时候，发给follower，follower放入unstable中
	每次leader的心跳或者app消息，就将大量的entry发给follower放入stable中
	 */
	ProgressSnapshot

	/*
	当snapshot备份完毕后，则变成replicate的状态，每次有数据来的时候再更新
	*/
	ProgressReplicate
)

//type progress struct {
//	Node map[string]ProgressState
//}

type ProgressState struct {
	Active   bool
	ActiveTS time.Time
	Type     int64
	Index    int64
	Term     int64
}
