package raft

// ReadOnly 结构体是用来确认entry已经到了follower上的
type ReadOnly struct {
	ReadOnlyQueue []string
	ReadOnlyMap   map[string]*ReadIndexStatus
}
type ReadIndexStatus struct {
	Acks      map[string]interface{} // ack的key是server的ip port信息
	Committed int64                  // 表示某个请求时候的committed值
	State     bool                   //是否已经响应过
}

func (r *ReadOnly) AddRequest(requestKey string, committed int64) {
	if _, ok := r.ReadOnlyMap[requestKey]; ok {
		return
	}
	r.ReadOnlyQueue = append(r.ReadOnlyQueue, requestKey)
	r.ReadOnlyMap[requestKey] = &ReadIndexStatus{
		Committed: committed,
		Acks:      make(map[string]interface{})}
}

// 收到端口的数据
func (r *ReadOnly) RecvAck(requestKey string, port string) {
	if _, ok := r.ReadOnlyMap[requestKey]; !ok {
		return
	}
	r.ReadOnlyMap[requestKey].Acks[port] = true
}

func NewReadOnly() *ReadOnly {
	q := make([]string, 0, 0)
	m := make(map[string]*ReadIndexStatus)
	r := ReadOnly{
		ReadOnlyMap:   m,
		ReadOnlyQueue: q,
	}
	return &r
}

// 移除一条数据
func (r *ReadOnly) RemoveOne(requestKey string) {
	_, ok := r.ReadOnlyMap[requestKey]
	if !ok {
		// TODO 错误
	}
	delete(r.ReadOnlyMap,requestKey)
	// TODO 将queue里面的数据清除
}

// 削减操作
