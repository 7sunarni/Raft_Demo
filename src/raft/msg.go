package raft

const (
	MsgNull = iota
	MsgHeartbeat
	MsgHeartBeatResp
	MsgApp
	MsgAppResp
	MsgAskVote
	MsgVoteResp
	// TODO
	MsgProp  // 提出建议：换leader?
	MsgSnap
	MsgUnreachable  // 节点不可到达
	MsgCheckQuorum  // 节点请求leader检查状态
	MsgReadIndex
	MsgReadIndexResp
)

type Msg struct {
	Type int64  `json:"type"`
	Data []byte `json:"data"`
	To   string `json:"to"`
	From string `json:"from"`
	Term int64  `json:"term"`
}

type Entry struct {
	Term        int64
	Index       int64
	Data        []byte
	Type        int8
	CommitCount int8
}

type Snapshot struct {
	Term         int64
	Index        int64
	Data         []byte
	SnapshotType int8
}

type RaftOperation struct {
	Operation string
	Key       string
	Value     int64
}
