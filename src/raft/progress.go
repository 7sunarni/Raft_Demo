package raft

/*
progress是描述以leader视角下的follower状态的结构体，
leader通过follower的progress状态来做出不同的响应：heartbeat,app,snapshot等
*/
type Progress struct {

}
