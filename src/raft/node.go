package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"raft/httpAPICenter"
	"raft/raftdebug"
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
	HeartBeat        = 5
	HeartBeatTimeout = HeartBeat + 10
)

type Node struct {
	Type int

	State int64
	Term  int64

	Quorum map[int64]map[string]interface{} // 用于检查是否满足投票条件。参数是周期，投票节点和投票结果

	OtherNode map[string]interface{} // 其他小伙伴的port

	progress map[string]ProgressState // 站在Leader视角下的Follower的状态

	DataIndex int64
	msgChan   chan Msg

	httpChan chan httpAPICenter.RaftOperation

	// 心跳超时
	heartBeatTimeoutTicker *time.Ticker

	// 选举超时
	electionTimeoutTicker time.Ticker

	Port string

	Log RaftLog

	Read *ReadOnly

	// debug日志，用于在前台展示
	*raftdebug.RaftDebugLog

	// 存放k-v的封装集合
	*ValueMap

	leader string
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
		progress:  make(map[string]ProgressState),
		ValueMap:  NewValueMap(),
	}
	return n
}

func (n *Node) Start() {
	n.init()
	if n.State == StateLeader {
		// 要先变为不是Leader状态才行，becomeLeader中有判断
		n.SetState(StateCandidate)
		n.becomeLeader()
	}
	rand.Seed(time.Now().UnixNano())
	// 移动位置，避免出现空指针问题
	duration := time.Duration(HeartBeatTimeout + rand.Int63n(5))
	n.heartBeatTimeoutTicker = time.NewTicker(duration * time.Second)
	go n.Monitor()
	n.RaftDebugLog.Warn(n.Port + " start")
	http.HandleFunc("/message", msgHandler)
	http.HandleFunc("/raft", raftHandler)
	http.HandleFunc("/debug", debugHandler)
	http.ListenAndServe(":"+n.Port, nil)
}

func (n *Node) SetState(state int64) {
	n.State = state
}

func (n *Node) init() {
	n.RaftDebugLog = raftdebug.NewRaftDebugLog()
	n.msgChan = make(chan Msg, 10)
	n.httpChan = make(chan httpAPICenter.RaftOperation, 10)
	n.Read = NewReadOnly()
}

func (n *Node) SetNodes(nodes []string) {
	for _, v := range nodes {
		n.OtherNode[v] = v
		//n.progress.Node[v] = ProgressState{
		//	Active: false,
		//	Type:   ProgressUnknown,
		//}
	}
}

// 用于输出debug的信息
func debugHandler(w http.ResponseWriter, r *http.Request) {
	// debug模块允许跨域请求
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json
	data, _ := ioutil.ReadAll(r.Body)
	if len(data) != 0 {
		w.Write([]byte("KILL"))
		os.Exit(1)
	}
	logs := n.RaftDebugLog.GetLogs()
	marshal, _ := json.Marshal(logs)
	w.Write(marshal)
}

// 用于对外部应用的CRUD api
func raftHandler(w http.ResponseWriter, r *http.Request) {
	data, e := ioutil.ReadAll(r.Body)
	if e != nil {
		n.RaftDebugLog.Error("raftHandler readAll error", e)
	}
	operation := httpAPICenter.RaftOperation{}
	if e := json.Unmarshal(data, &operation); e != nil {
		n.RaftDebugLog.Error("unmarshal error", e)
	}
	requestKey := n.Port +
		strconv.Itoa(time.Now().Hour()) +
		strconv.Itoa(time.Now().Minute()) +
		strconv.Itoa(time.Now().Second())
	operation.RequestKey = requestKey
	operation.Port = n.Port
	// 生成requestKey之后返回到data里面去
	data, _ = json.Marshal(operation)
	msg := Msg{
		From: n.Port,
		To:   n.Port,
		Data: data,
		Type: MsgReadIndex,
	}
	n.RaftDebugLog.Info("node rcv operation")
	n.RaftDebugLog.Info(msg)
	n.msgChan <- msg
	var respMsg httpAPICenter.RaftOperation
	httpTimeoutTicker := time.NewTicker(20 * time.Second)
forLoop:
	for {
		select {
		case respMsg = <-n.httpChan:
			break forLoop
		case <-httpTimeoutTicker.C:
			break forLoop
		}
	}
	n.RaftDebugLog.Info("rcv httpChan", requestKey, respMsg.RequestKey)
	if requestKey == respMsg.RequestKey {
		marshal, _ := json.Marshal(respMsg)
		w.Write(marshal)
	} else {
		w.Write([]byte("TIMEOUT"))
	}
}

// 用于raft节点之间通信的api
func msgHandler(w http.ResponseWriter, r *http.Request) {
	data, e := ioutil.ReadAll(r.Body)
	if e != nil {
		n.RaftDebugLog.Error("handler read error", e)
	}
	msg := Msg{}
	if e := json.Unmarshal(data, &msg); e != nil {
		n.RaftDebugLog.Error("unmarshal error", e)
	}
	if msg.Type == MsgVoteResp {
		n.RaftDebugLog.Error("recv vote resp", msg)
	}
	n.msgChan <- msg
	w.Write(nil)
}

func (n *Node) MsgSender(msg Msg) {
	defer func() {
		if e := recover(); e != nil {
			n.RaftDebugLog.Fatal("msg sender recover error", e)
		}
	}()
	n.RaftDebugLog.Trace("send msg", msg.Type, msg.To)
	// TODO 不用修改为全局的HttpClient，后期会改为gRPC方式用于节点间通信
	cli := http.Client{}
	msgReader, _ := json.Marshal(msg)
	reader := bytes.NewReader(msgReader)
	_, err := cli.Post("http://localhost:"+msg.To+"/message", "", reader)

	if _, isOK := n.OtherNode[msg.To]; isOK && err != nil {
		delete(n.OtherNode, msg.To)
	}
	//// 其他属性不变，active变成false
	//n.Node[msg.To] = ProgressState{
	//	Active: false,
	//	Type:   n.Node[msg.To].Type,
	//}
}

func (n *Node) Monitor() {
	defer func() {
		if e := recover(); e != nil {
			n.RaftDebugLog.Error("monitor recover err: ", e)
			debug.PrintStack()
		}
	}()
	for {
		select {
		case <-n.heartBeatTimeoutTicker.C:
			if n.heartBeatTimeoutTicker != nil {
				n.heartBeatTimeoutTicker.Stop()
			}
			if n.State == StateFollower {
				n.RaftDebugLog.Info("heartbeat time out, start new election term")
				n.startElection(false)
			}
		case msg := <-n.msgChan:
			n.MsgHandler(msg)
		}
	}
}
func (n *Node) startElection(voted bool) {
	if n.State != StateFollower {
		return
	}
	n.SetState(StateCandidate)
	// TODO 增加选举超时，如果投票没投自己就取消选举
	n.heartBeatTimeoutTicker.Stop()
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
	n.Visit(msg)
}

func (n *Node) MsgHandler(msg Msg) {
	defer func() {
		if e := recover(); e != nil {
			n.RaftDebugLog.Error("msg handler recover", e)
			debug.PrintStack()
		}
	}()
	switch msg.Type {
	case MsgNull:
		break
	case MsgHeartbeat:
		n.State = StateFollower
		// 更新当前节点的leader节点
		n.leader = msg.From
		rand.Seed(time.Now().UnixNano())
		duration := time.Duration(HeartBeatTimeout + rand.Int63n(5))
		n.heartBeatTimeoutTicker = time.NewTicker(duration * time.Second)
		if n.State != StateLeader {
			// 这里是否拒绝
			committed := msg.Index
			n.RaftDebugLog.Trace("follower recv heartbeat ", &committed == nil, committed)
			if committed >= 100 {
				n.RaftDebugLog.Info("follower get read index heart beat")
				msg.Reject = n.Log.Commit(committed - 100)
			} else {
				var otherNode map[string]interface{}
				json.Unmarshal(msg.Data, &otherNode)
				n.OtherNode = otherNode
				//var p progress
				//json.Unmarshal(msg.Data, &p)
				//n.progress = p
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
			n.RaftDebugLog.Info("leader get read index heart beat resp")
			n.Read.RecvAck(string(msg.Data), msg.From)
			status := n.Read.ReadOnlyMap[string(msg.Data)]
			if n.checkQuorum(status.Acks) && status.State == false {
			}
		}
		break
	case MsgApp:
		if n.State == StateLeader {
			n.RaftDebugLog.Info("leader rcv MsgApp")
			term, index := n.Log.LastIndexAndTerm()
			// Leader添加肯定会成功
			e := Entry{
				Term:  term,
				Index: index + 1,
				Data:  msg.Data,
			}
			operation := httpAPICenter.RaftOperation{}
			json.Unmarshal(msg.Data, &operation)
			// TODO error处理
			if operation.Operation == OperationAdd {
				n.AddValue(operation.Key, operation.Value)
			}
			if operation.Operation == OperationUpdate {
				n.UpdateValue(operation.Key, operation.Value)
			}
			if operation.Operation == OperationGet {
				value, _ := n.GetValue(operation.Key)
				operation.Value = value
			}
			n.Log.AppendEntry(e.Term, e.Index, e)
			// Leader收到MsgApp，首先自己添加一条请求，然后自己确认请求
			n.Read.AddRequest(string(e.Index), n.Log.Committed, operation.Value)
			n.Read.RecvAck(string(e.Index), n.Port)
			n.RaftDebugLog.Trace("=== leader log ", n.Log)
			n.RaftDebugLog.Info("=== leader map ", n.Map)
			bytesData, err := json.Marshal(e)
			if err != nil {
				n.RaftDebugLog.Error("marshal entry error", e)
			}
			n.RaftDebugLog.Warn(n.Port + " rcv App as leader")
			broadCastMsgApp := Msg{
				Type: MsgApp,
				From: n.Port,
				Data: bytesData,
			}
			n.Visit(broadCastMsgApp)
		} else {
			e := Entry{}
			json.Unmarshal(msg.Data, &e)
			// 写入map中
			operation := httpAPICenter.RaftOperation{}
			json.Unmarshal(e.Data, &operation)
			if operation.Operation == OperationAdd {
				n.AddValue(operation.Key, operation.Value)
			}
			if operation.Operation == OperationUpdate {
				n.UpdateValue(operation.Key, operation.Value)
			}
			n.RaftDebugLog.Info("follower rcv MsgApp", operation)
			isOk, t, i2 := n.Log.AppendEntry(e.Term, e.Index, e)
			n.RaftDebugLog.Trace("=== follower log ", n.Log)
			n.RaftDebugLog.Info("=== follower map ", n.Map)
			n.RaftDebugLog.Info(n.Port + " rcv App as follower")
			broadCastMsgAppResp := Msg{
				Type:   MsgAppResp,
				From:   n.Port,
				To:     msg.From,
				Term:   t,
				Index:  i2,
				Reject: isOk,
				Data:   msg.Data,
			}
			go n.MsgSender(broadCastMsgAppResp)
		}
	case MsgAppResp:
		n.RaftDebugLog.Info("leader rcv AppResp", msg.From)
		n.Read.RecvAck(string(msg.Index), msg.From)
		status := n.Read.ReadOnlyMap[string(msg.Index)]
		n.RaftDebugLog.Info(n.Read.ReadOnlyMap)
		if n.checkQuorum(status.Acks) && status.State == false {
			n.RaftDebugLog.Warn("AppResp ====== leader get committed ok ======", len(status.Acks), 1+(len(n.OtherNode))/2)
			e := Entry{}
			json.Unmarshal(msg.Data, &e)
			backOperation := httpAPICenter.RaftOperation{}
			json.Unmarshal(e.Data, &backOperation)
			n.RaftDebugLog.Warn("unmarshal", backOperation)
			operation := httpAPICenter.RaftOperation{
				Value:      status.TempValue,
				RequestKey: backOperation.RequestKey,
				Port:       backOperation.Port,
			}
			n.RaftDebugLog.Warn(operation)
			n.dealCommittedOperation(operation)
			status.State = true
		}
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
				voteMsg.Data = []byte(VoteReject)
				voted = false
			} else {
				voteMsg.Data = []byte(VoteAccepted)
				voted = true
			}
			go n.MsgSender(voteMsg)
			go n.startElection(voted)
		}
		break
	case MsgVoteResp:
		n.Quorum[n.Term][msg.From] = true
		n.RaftDebugLog.Trace("handle vote resp")
		n.RaftDebugLog.Trace(n.Term, n.Quorum)
		if n.checkQuorum(n.Quorum[n.Term]) {
			n.becomeLeader()
		}
		break
	case MsgReadIndex:
		// 如果不是Leader收到请求转发给Leader处理请求
		if n.Type == StateFollower || n.Type == StateCandidate {
			msg.From = n.Port
			msg.To = n.leader
			go n.MsgSender(msg)
			n.RaftDebugLog.Info("send MsgReadIndex to leader", msg.To)
		}
		if n.State == StateLeader {
			n.RaftDebugLog.Info("leader rcv MsgReadIndex msg", msg)
			msg.Type = MsgApp
			n.msgChan <- msg
		}
		break
	case MsgReadIndexResp:
		operation := httpAPICenter.RaftOperation{}
		json.Unmarshal(msg.Data, &operation)
		n.RaftDebugLog.Warn("rcv MsgReadIndexResp", msg.From, msg.To, operation)
		n.httpChan <- operation
	}

}

func (n *Node) heartBeatTicker() {
	defer func() {
		if e := recover(); e != nil {
			n.RaftDebugLog.Fatal("recover error", e)
		}
	}()
	n.OtherNode[n.Port] = n.Port
	heartBeat := time.NewTicker(HeartBeat * time.Second)
	for {
		select {
		case <-heartBeat.C:
			d := n.OtherNode
			//d := n.progress
			data, e := json.Marshal(d)
			if e != nil {
				n.RaftDebugLog.Error("data error")
			}
			msg := Msg{
				Type: MsgHeartbeat,
				From: n.Port,
				Data: data,
				Term: n.Term,
			}
			n.Visit(msg)
		}
	}
}

// 用于校验是否通过大部分节点，通用方法
func (n *Node) checkQuorum(m map[string]interface{}) bool {
	n.RaftDebugLog.Info("check quorum ", len(m), len(n.OtherNode)/2+1)
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
	n.RaftDebugLog.Warn(fmt.Sprintf("===== %v become leader =====", n.Port))
	go n.heartBeatTicker()
}

func (n *Node) changeLeader() {

}

// 消息确认后返回给请求发起的节点
func (n *Node) dealCommittedOperation(operation httpAPICenter.RaftOperation) {
	data, _ := json.Marshal(operation)
	msg := Msg{
		Type: MsgReadIndexResp,
		From: n.Port,
		To:   operation.Port,
		Data: data,
	}
	n.RaftDebugLog.Warn("deal operation", msg.To, msg.From, operation)
	if n.Port == operation.Port {
		n.msgChan <- msg
	} else {
		go n.MsgSender(msg)
	}

}

// visit函数用于向所有的节点发送一条消息，
// 比如心跳、超时选举、更新entry等操作
func (n *Node) Visit(msg Msg) {
	for p := range n.OtherNode {
		if p == n.Port {
			continue
		}
		msg.To = p
		go n.MsgSender(msg)
	}
	// 更新操作
	//for item := range n.progress.Node {
	//	if item == n.Port {
	//		continue
	//	}
	//	msg.To = item
	//	go n.MsgSender(msg)
	//}
}
