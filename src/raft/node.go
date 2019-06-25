package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

const (
	StateFollower = iota
	StateCandidate
	StateLeader
)
const (
	TestHeartBeatTimeout = 20
	TestHeartBeat        = TestHeartBeatTimeout - 15
)

type Node struct {
	Type int64

	State int64
	Term  int64
	Server
	Quorum int64

	OtherNode map[string]string //其他小伙伴的port

	DataIndex int64
	MsgChan   chan Msg

	// 心跳超时
	HeartBeatTimeoutTicker *time.Ticker

	// 选举超时
	ElectionTimeoutTicker time.Ticker

	Port string

	Log RaftLog
}

var n *Node

func NewNode(port string) *Node {
	stableLog := StableLog{FileName: TestLogFilePrefix + port + TestLogFileSuffix}
	raftLog := RaftLog{
		Stable: stableLog,
	}
	n = &Node{
		Port:      port,
		OtherNode: make(map[string]string),
		Log:       raftLog,
	}
	return n
}

func (n *Node) SetNodes(nodes []string) {
	for _, v := range nodes {
		n.OtherNode[v] = v
	}
}

func (n *Node) SetState(state int64) {
	n.State = state
}

func (n *Node) Start() {
	if n.State == StateLeader {
		go n.HeartBeatStart()
	}
	n.MsgChan = make(chan Msg, 10000)
	go n.Monitor()
	// 20秒随机两秒方便测试同时超时
	n.HeartBeatTimeoutTicker = time.NewTicker(time.Duration(TestHeartBeatTimeout+rand.Int63n(8)) * time.Second)
	//n.HeartBeatTimeoutTicker = time.NewTicker(20 * time.Second)
	http.HandleFunc("/message", MsgHandler)
	http.HandleFunc("/raft", raftHandler)
	fmt.Println(n.Port + " start")
	http.ListenAndServe(":"+n.Port, nil)
}

func raftHandler(w http.ResponseWriter, r *http.Request) {
	data, e := ioutil.ReadAll(r.Body)
	if e != nil {
		fmt.Println("raftHandler readAll error", e)
	}
	opera := RaftOperation{}
	if e := json.Unmarshal(data, &opera); e != nil {
		fmt.Println("unmarshal error", e)
	}
	fmt.Println(opera)
}

func MsgHandler(w http.ResponseWriter, r *http.Request) {
	data, e := ioutil.ReadAll(r.Body)
	if e != nil {
		fmt.Println("handler read error", e)
	}
	msg := Msg{}
	if e := json.Unmarshal(data, &msg); e != nil {
		fmt.Println("unmarshal error", e)
	}
	if msg.Type == MsgVoteResp {
		fmt.Println("recv vote resp", msg)
		//v, s := <-n.MsgChan
		//fmt.Println(v, s)
	}
	n.MsgChan <- msg
	w.Write(nil)
}

func (n *Node) MsgSender(msg Msg) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("msg sender recover error", e)
		}
	}()
	fmt.Println("send msg", msg.Type, msg.To)
	cli := http.Client{}
	msgReader, e := json.Marshal(msg)
	if e != nil {
		fmt.Println("data error", e)
	}
	reader := bytes.NewReader(msgReader)
	go func() {
		_, e = cli.Post("http://localhost:"+msg.To+"/message", "", reader)
		if e != nil {
			delete(n.OtherNode, msg.To)
		}
	}()
	return
}

func (n *Node) Monitor() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("monitor recover err: ", e)
		}
	}()
	for {
		select {
		case c, ok := <-n.HeartBeatTimeoutTicker.C:
			if n.HeartBeatTimeoutTicker != nil {
				n.HeartBeatTimeoutTicker.Stop()
			}
			fmt.Println("timeoutTicker", c, ok)
			if n.State == StateFollower {
				go n.startElection()
				fmt.Println("heartbeat time out,start new election term")
			}
		case msg := <-n.MsgChan:
			go n.MsgHandler(msg)
		}
	}
}
func (n *Node) startElection() {
	if n.State != StateFollower {
		return
	}
	n.State = StateCandidate
	// 取消心跳 TODO 增加选举周期超时
	n.HeartBeatTimeoutTicker.Stop()
	//n.HeartBeatTimeoutTicker = nil
	n.Term++
	// 先投一票给自己
	n.Quorum = 1
	msg := Msg{
		Type: MsgAskVote,
		From: n.Port,
		Data: nil,
		Term: n.Term,
	}
	n.visit(msg)
}

func (n *Node) MsgHandler(msg Msg) {
	if msg.Type == MsgVoteResp {
		fmt.Println("handle vote resp")
	}
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("MsgHandler recover error", e)
		}
	}()
	switch msg.Type {
	case MsgNull:
		break
	case MsgHeartbeat:
		n.State = StateFollower
		// TODO 封装成函数
		n.HeartBeatTimeoutTicker = time.NewTicker(time.Duration(TestHeartBeatTimeout+rand.Int63n(8)) * time.Second)
		if n.State != StateLeader {
			var otherNode map[string]string
			json.Unmarshal(msg.Data, &otherNode)
			n.OtherNode = otherNode
			msg.Type = MsgHeartBeatResp
			temp := msg.From
			msg.From = msg.To
			msg.To = temp
			data, _ := json.Marshal("HeartBeatResp")
			msg.Data = data
			go n.MsgSender(msg)
			fmt.Println(n.Port, "HeartBeat", n.OtherNode, n.Term)
		}
		break
	case MsgHeartBeatResp:
		if n.State == StateLeader {
			fmt.Println("heartbeat resp")
		}
		break
	case MsgAskVote:
		if n.State == StateFollower {
			// TODO  判断选举周期 msg.Term > n.Term
			voteMsg := Msg{
				Type: MsgVoteResp,
				To:   msg.From,
				From: msg.To,
				Term: n.Term,
			}
			go n.MsgSender(voteMsg)
			go n.startElection()
		}
		break
	case MsgVoteResp:
		fmt.Println("recv vote resp", n.Quorum)
		n.Quorum++
		if n.checkQuorum() {
			n.becomeLeader()
		}
		break
	}
}

func (n *Node) HeartBeatStart() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("recover error", e)
		}
	}()
	n.OtherNode[n.Port] = n.Port
	heartBeat := time.NewTicker(TestHeartBeat * time.Second)
	for {
		select {
		case <-heartBeat.C:
			d := n.OtherNode
			data, e := json.Marshal(d)
			if e != nil {
				fmt.Println("data error")
			}
			msg := Msg{
				Type: MsgHeartbeat,
				From: n.Port,
				Data: data,
				Term: n.Term,
			}
			n.visit(msg)
		}
	}
}

func (n *Node) checkQuorum() bool {
	b := n.Quorum > int64((len(n.OtherNode)-1)/2)
	fmt.Println("n.Quorum", len(n.OtherNode), b)
	return b
}

func (n *Node) becomeLeader() {
	if n.State == StateLeader {
		return
	}
	n.State = StateLeader
	fmt.Println("===== become leader === ", n.Port)
	go n.HeartBeatStart()
}

func (n *Node) changeLeader() {

}

// visit函数用于向所有的节点发送一条消息，
// 比如心跳、超时选举、更新entry等操作
func (n *Node) visit(msg Msg) {
	for p := range n.OtherNode {
		if p == n.Port {
			continue
		}
		msg.To = p
		go n.MsgSender(msg)
	}
}
