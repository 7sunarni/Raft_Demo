package raft

const (
	MsgNull = iota
	MsgHeartbeat
	MsgHeartBeatResp
	MsgAskVote
	MsgVoteResp
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
