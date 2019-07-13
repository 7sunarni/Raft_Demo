package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

const (
	StateFollower = iota
	StateCandidate
	StateLeader
)
const (
	TestHeartBeat        = 5
	TestHeartBeatTimeout = TestHeartBeat + 10
)

type Node struct {
	Type int

	State int64
	Term  int64

	Quorum map[int64]map[string]interface{} // 用于检查是否满足投票条件

	OtherNode map[string]interface{} //其他小伙伴的port

	DataIndex int64
	MsgChan   chan Msg

	ReadChan chan Msg

	// 心跳超时
	HeartBeatTimeoutTicker *time.Ticker

	// 选举超时
	ElectionTimeoutTicker time.Ticker

	Port string

	Log RaftLog

	Read *ReadOnly
}

var n *Node

func NewNode(port string) *Node {
	stableLog := StableLog{FileName: TestLogFilePrefix + port + TestLogFileSuffix}
	raftLog := RaftLog{
		Stable: stableLog,
	}
	n = &Node{
		Port:      port,
		OtherNode: make(map[string]interface{}),
		Log:       raftLog,
		Quorum:    make(map[int64]map[string]interface{}),
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
	n.MsgChan = make(chan Msg, 10)
	n.ReadChan = make(chan Msg, 10)
	n.Read = NewReadOnly()
	rand.Seed(time.Now().UnixNano())
	// 移动位置，避免出现空指针问题
	duration := time.Duration(TestHeartBeatTimeout + rand.Int63n(5))
	n.HeartBeatTimeoutTicker = time.NewTicker(duration * time.Second)
	go n.Monitor()
	http.HandleFunc("/message", MsgHandler)
	http.HandleFunc("/raft", raftHandler)
	http.HandleFunc("/read", readHandler)
	fmt.Println(n.Port + " start")
	http.ListenAndServe(":"+n.Port, nil)
}

// 用于暴露对外read的api
func readHandler(w http.ResponseWriter, r *http.Request) {
	data, e := ioutil.ReadAll(r.Body)
	if e != nil {
		fmt.Println("raftHandler readAll error", e)
	}
	opera := RaftOperation{}
	if e := json.Unmarshal(data, &opera); e != nil {
		fmt.Println("unmarshal error", e)
	}
	requestKey := n.Port +
		strconv.Itoa(time.Now().Hour()) +
		strconv.Itoa(time.Now().Minute()) +
		strconv.Itoa(time.Now().Second())
	fmt.Println("generate requestKey", requestKey)

	msg := Msg{
		Type: MsgReadIndex,
		From: n.Port,
		To:   n.Port,
		Data: []byte(requestKey),
	}
	n.MsgChan <- msg
	resp := "NODATA"
	httpTimeoutTicker := time.NewTicker(20 * time.Second)

forLoop:
	for {
		select {
		case msg := <-n.ReadChan:
			resp = string(msg.Data)
			fmt.Println("recv requestKey", resp)
			break forLoop
		case <-httpTimeoutTicker.C:
			break forLoop
		}
	}
	w.Write([]byte(resp))
}

// 用于暴露对外的api
func raftHandler(w http.ResponseWriter, r *http.Request) {
	data, e := ioutil.ReadAll(r.Body)
	if e != nil {
		fmt.Println("raftHandler readAll error", e)
	}
	opera := RaftOperation{}
	if e := json.Unmarshal(data, &opera); e != nil {
		fmt.Println("unmarshal error", e)
	}
	msg := Msg{
		Type: MsgApp,
		From: n.Port,
		To:   n.Port,
	}
	n.MsgChan <- msg
	w.Write(nil)
}

// 用于raft节点之间通信的api
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
	msgReader, _ := json.Marshal(msg)
	reader := bytes.NewReader(msgReader)
	_, err := cli.Post("http://localhost:"+msg.To+"/message", "", reader)
	if err != nil {
		delete(n.OtherNode, msg.To)
	}
}

func (n *Node) Monitor() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("monitor recover err: ", e)
			debug.PrintStack()
		}
	}()
	for {
		select {
		case c, _ := <-n.HeartBeatTimeoutTicker.C:
			if n.HeartBeatTimeoutTicker != nil {
				n.HeartBeatTimeoutTicker.Stop()
			}
			{
				fmt.Printf("=== timeoutTicker chan  %v \n", c, )
				fmt.Printf("=== timeoutTicker ticker %v \n", n.HeartBeatTimeoutTicker)
				fmt.Println("time now", time.Now())
			}
			if n.State == StateFollower {
				fmt.Println("heartbeat time out, start new election term")
				go n.startElection(false)
			}
		case msg := <-n.MsgChan:
			go n.MsgHandler(msg)
		}
	}
}
func (n *Node) startElection(voted bool) {
	if n.State != StateFollower {
		return
	}
	n.State = StateCandidate
	// TODO 增加选举超时，如果投票没投自己就取消选举
	n.HeartBeatTimeoutTicker.Stop()
	n.Term++
	// 如果已经投过票，就不投票直接开始返回
	if voted {
		return
	}
	if _, ok := n.Quorum[n.Term]; !ok {
		n.Quorum[n.Term] = make(map[string]interface{})
	}
	n.Quorum[n.Term][n.Port] = true
	msg := Msg{
		Type: MsgAskVote,
		From: n.Port,
		Data: nil,
		Term: n.Term,
	}
	n.visit(msg)
}

func (n *Node) MsgHandler(msg Msg) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("MsgHandler recover error", e)
			debug.PrintStack()
		}
	}()
	switch msg.Type {
	case MsgNull:
		break
	case MsgHeartbeat:
		n.State = StateFollower
		rand.Seed(time.Now().UnixNano())
		duration := time.Duration(TestHeartBeatTimeout + rand.Int63n(5))
		n.HeartBeatTimeoutTicker = time.NewTicker(duration * time.Second)
		{
			fmt.Printf("=== update hb ticker %v \n", n.HeartBeatTimeoutTicker)
			fmt.Printf("=== update hb ticker duration %v \n", duration)
			fmt.Printf("=== update hb ticker now  %v \n", time.Now())
		}
		if n.State != StateLeader {
			// 这里是否拒绝
			committed := msg.Index
			fmt.Println("follower recv heartbeat ", &committed == nil, committed)
			if committed >= 100 {
				fmt.Println("follower get read index heart beat")
				msg.Reject = n.Log.Commit(committed - 100)
			} else {
				var otherNode map[string]interface{}
				json.Unmarshal(msg.Data, &otherNode)
				n.OtherNode = otherNode
			}
			msg.Type = MsgHeartBeatResp
			temp := msg.From
			msg.From = msg.To
			msg.To = temp
			go n.MsgSender(msg)
		}
		break
	case MsgHeartBeatResp:
		if n.State != StateLeader {
			break
		}
		committed := msg.Index
		if committed >= 100 && msg.Reject == false {
			// 心跳返回正确后，更新readonly的数据
			fmt.Println("leader get read index heart beat resp")
			n.Read.RecvAck(string(msg.Data), msg.From)
			status := n.Read.ReadOnlyMap[string(msg.Data)]
			acks := len(status.Acks)
			if acks > 1+(len(n.OtherNode))/2 && status.State == false {
				fmt.Println("====== leader get committed ok ======", acks, 1+(len(n.OtherNode))/2)
				n.ReadChan <- msg
				status.State = true
			}
		}
		break
	case MsgApp:
		if n.State == StateLeader {
			index, term := n.Log.LastIndexAndTerm()
			// Leader添加肯定会成功
			e := Entry{
				Term:  term + 1,
				Index: index + 1,
			}
			n.Log.AppendEntry(term, index, e)
			fmt.Println("=== leader log ", n.Log)
			bytes, err := json.Marshal(e)
			if err != nil {
				fmt.Println("marshal entry error", e)
			}
			fmt.Println(n.Port + " rcv App as leader")
			broadCastMsgApp := Msg{
				Type: MsgApp,
				From: n.Port,
				Data: bytes,
			}
			n.visit(broadCastMsgApp)
		} else {
			e := Entry{}
			json.Unmarshal(msg.Data, &e)
			isOk, t, i2 := n.Log.AppendEntry(e.Term-1, e.Index-1, e)
			fmt.Println("=== follower log ", n.Log)
			fmt.Println(n.Port + " rcv App as follower")
			broadCastMsgAppResp := Msg{
				Type:   MsgAppResp,
				From:   n.Port,
				To:     msg.From,
				Term:   t,
				Index:  i2,
				Reject: isOk,
			}
			go n.MsgSender(broadCastMsgAppResp)
		}
	case MsgAppResp:
		// LEADER对相应的数据进行处理
		fmt.Println(n.Port, "rcv msg app resp")
	case MsgAskVote:
		if n.State == StateFollower {
			voteMsg := Msg{
				Type: MsgVoteResp,
				To:   msg.From,
				From: msg.To,
				Term: n.Term,
			}
			var voted bool
			if msg.Term < n.Term {
				voteMsg.Data = []byte("REJECT")
				voted = false
			} else {
				voteMsg.Data = []byte("VOTE")
				voted = true
			}
			go n.MsgSender(voteMsg)
			go n.startElection(voted)
		}
		break
	case MsgVoteResp:
		n.Quorum[n.Term][msg.From] = true
		fmt.Println("handle vote resp")
		fmt.Println(n.Term, n.Quorum)
		if n.checkQuorum(n.Quorum[n.Term]) {
			n.becomeLeader()
		}
		break
	case MsgReadIndex:
		if n.State == StateLeader {
			requestKey := string(msg.Data)
			n.Read.AddRequest(requestKey, n.Log.Committed)
			n.Read.RecvAck(requestKey, n.Port)
			msg := Msg{
				Type:  MsgHeartbeat,
				From:  n.Port,
				Data:  []byte(requestKey),
				Term:  n.Term,
				Index: n.Log.Committed + 100,
			}
			fmt.Println("leader send read index heart beat")
			n.visit(msg)
		}
		if n.Type == StateFollower {
			// 转发给leader处理
		}
		if n.Type == StateCandidate {
			// 转发给leader处理
		}
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

// 用于校验是否通过大部分节点，通用方法
func (n *Node) checkQuorum(m map[string]interface{}) bool {
	fmt.Print("check quorum ", len(m), len(n.OtherNode)/2+1)
	if len(m) >= len(n.OtherNode)/2+1 {
		return true
	}
	return false
}

func (n *Node) becomeLeader() {
	if n.State == StateLeader {
		return
	}
	n.State = StateLeader
	fmt.Printf("===== %v become leader === \n", n.Port)
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
