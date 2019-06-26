package raft

// RaftLog中的持久化储存和不稳定储存
type LogStorage interface {
	FirstIndex() int64      // 获取第一条索引
	LastIndex() int64       // 获取最后一条索引
	Term(index int64) int64 // 获取索引所在的周期
}

type RaftLog struct {
	Stable   StableLog
	Unstable UnstableLog
	// Committed提交表示已经写入到Unstable中的索引
	// 每次append entry到stable中不一定会成功
	Committed int64
}

// 向LOG中添加日志，添加成功后返回true、添加的index和term
// 添加失败后返回false、当前LOG中的最大index和term
func (r *RaftLog) AppendEntry(index, term int64, entries ...Entry) (isOk bool, i int64, t int64) {
	isMatch, err := r.MatchTerm(index, term)
	if err != nil || !isMatch {
		lastIndex, lastTerm := r.LastIndexAndTerm()
		return false, lastIndex, lastTerm
	}
	r.Unstable.AppendEntry(entries...)
	r.Committed += int64(len(entries))
	return true, index, term
}

// 向LOG中添加新周期的日志，添加成功后返回true、添加的index和term
// 添加失败后返回false、当前LOG中的最大index和term
func (r *RaftLog) NewTermAppendEntry(index, term int64, entries ...Entry) (isOk bool, i int64, t int64) {
	return r.AppendEntry(index, term-1, entries...)
}

// 向LOG中添加快照
func (r *RaftLog) AppendSnapshot(index, term int64, snapshot Snapshot) {
	r.Unstable.AppendSnapshot(snapshot)
}

func (r *RaftLog) MatchTerm(index, term int64) (isMatch bool, err error) {
	var firstIndex int64
	var lastIndex int64
	if firstIndex = r.Stable.FirstIndex(); firstIndex == 0 {
		firstIndex = r.Unstable.FirstIndex()
	}

	if lastIndex = r.Unstable.LastIndex(); lastIndex == 0 {
		lastIndex = r.Stable.LastIndex()
	}

	t := r.Term(index - 1)
	if t == term {
		return true, nil
	}

	return false, nil
}

func (r *RaftLog) Term(index int64) int64 {
	if t := r.Unstable.Term(index); t != 0 {
		return t
	}
	if t := r.Stable.Term(index); t != 0 {
		return t
	}
	return 0
}

// 找到和当前entry冲突的LOG的周期
func (r *RaftLog) FindConflict(entry Entry) int64 {
	if &r.Stable != nil &&
		r.Stable.LastIndex() == entry.Index &&
		r.Stable.Term(entry.Index) != entry.Term {
		return r.Stable.Term(entry.Index)
	}
	if r.Unstable.Snapshot != nil &&
		r.Unstable.Snapshot.Index == entry.Index &&
		r.Unstable.Snapshot.Term != entry.Term {
		return r.Unstable.Snapshot.Term
	}
	for _, e := range r.Unstable.Entries {
		if e.Index == entry.Index && e.Term != entry.Term {
			return e.Term
		}
	}
	// 考虑返回error
	return 0
}

// 找到当前LOG中的最大的Index和Term
func (r *RaftLog) LastIndexAndTerm() (index, term int64) {
	var lastIndex int64
	if lastIndex = r.Unstable.LastIndex(); lastIndex == 0 {
		lastIndex = r.Stable.LastIndex()
	}
	if len(r.Unstable.Entries) == 0 {
		return 0, 0
	}
	lastTerm := r.Unstable.Entries[len(r.Unstable.Entries)-1].Term
	return lastIndex, lastTerm
}
