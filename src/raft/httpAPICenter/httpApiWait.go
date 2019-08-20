package httpAPICenter

import (
	"encoding/json"
	"net/http"
	"raft/raftdebug"
	"sync"
	"time"
)

type HttpApiWait struct {
	Timestamp  int64
	RequestKey string
}

type HttpApi struct {
	RespMap     map[HttpApiWait]*http.ResponseWriter
	CheckTicker *time.Ticker
	HttpChan    chan RaftOperation
	m           sync.Mutex
	DebugLog    *raftdebug.RaftDebugLog
}

// CRUD API
type RaftOperation struct {
	Operation  string
	Key        string
	Value      int64
	RequestKey string // 用于在服务内部生成key调用
}

func NewHttpApi() *HttpApi {
	httpApi := &HttpApi{
		RespMap:     make(map[HttpApiWait]*http.ResponseWriter),
		CheckTicker: time.NewTicker(time.Second),
		HttpChan:    make(chan RaftOperation, 10),
		m:           sync.Mutex{},
	}
	return httpApi
}

func (h *HttpApi) Run() {
	h.DebugLog.Warn("httpApi Running")
	defer func() {
		if e := recover(); e != nil {
			h.DebugLog.Fatal("HttpApi Run", e)
		}
	}()
	for {
		select {
		case <-h.CheckTicker.C:
			h.clearMap(time.Now().Unix())
			break
		case raftOperation := <-h.HttpChan:
			h.DebugLog.Warn("httpApi get operation from httpChan", raftOperation, h.RespMap)
			h.httpResp(raftOperation)
			break
		}
	}
}

func (h *HttpApi) httpResp(raftOperation RaftOperation) {
	for key := range h.RespMap {
		if raftOperation.RequestKey == key.RequestKey {
			w := *h.RespMap[key]
			bytes, _ := json.Marshal(raftOperation)
			h.DebugLog.Warn("response", raftOperation)
			i, e := w.Write(bytes)
			h.DebugLog.Warn("i", i, "error", e)
			delete(h.RespMap, key)
		}
	}
}

func (h *HttpApi) clearMap(nowTimeStamp int64) {
	for key := range h.RespMap {
		if nowTimeStamp-key.Timestamp > 20 {
			w := *h.RespMap[key]
			w.Write([]byte("TIMEOUT"))
			h.DebugLog.Warn("TIMEOUT")
			delete(h.RespMap, key)
		}
	}
}

func (h *HttpApi) AddNewResponse(w *http.ResponseWriter, requestKey string) {
	h.DebugLog.Warn("recv new response writer")
	h.RespMap[HttpApiWait{Timestamp: time.Now().Unix(), RequestKey: requestKey}] = w
}
