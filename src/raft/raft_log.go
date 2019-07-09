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
	// Committed的值是 len(entry)+1 = LastEntry.Index + 1，因为entry的Index索引从0开始
	Committed int64
}

// 向LOG中添加日志，添加成功后返回true、添加的index和term
// 返回数据：
//  - 成功 返回 true, 添加后的Term，添加后的Index
//  - 失败 返回 false，当前Log最大的Term和当前Log最大的Term
func (r *RaftLog) AppendEntry(term, index int64, entries ...Entry) (isOk bool, t int64, i int64) {
	isMatch, err := r.MatchTerm(term, index)
	if err != nil || !isMatch {
		lastIndex, lastTerm := r.LastIndexAndTerm()
		return false, lastTerm, lastIndex
	}
	r.Unstable.AppendEntry(entries...)
	// TODO 这里要判断下
	r.Committed += int64(len(entries))
	return true, entries[len(entries)-1].Term, entries[len(entries)-1].Index
}

// 向LOG中添加新周期的日志，添加成功后返回true、添加的index和term
// 添加失败后返回false、当前LOG中的最大index和term
func (r *RaftLog) NewTermAppendEntry(term, index int64, entries ...Entry) (isOk bool, i int64, t int64) {
	return r.AppendEntry(term-1, index, entries...)
}

// 向LOG中添加快照
func (r *RaftLog) AppendSnapshot(term, index int64, snapshot Snapshot) {
	r.Unstable.AppendSnapshot(snapshot)
}

func (r *RaftLog) MatchTerm(term, index int64) (isMatch bool, err error) {
	var firstIndex int64
	var lastIndex int64
	if firstIndex = r.Stable.FirstIndex(); firstIndex == 0 {
		firstIndex = r.Unstable.FirstIndex()
	}

	if lastIndex = r.Unstable.LastIndex(); lastIndex == 0 {
		lastIndex = r.Stable.LastIndex()
	}

	t := r.Term(index)
	if t == term {
		return true, nil
	}

	return false, nil
}

func (r *RaftLog) Term(index int64) int64 {
	if t := r.Unstable.Term(index); t != -1 {
		return t
	}
	if t := r.Stable.Term(index); t != -1 {
		return t
	}
	return -1
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

// 找到当前LOG中的最大的Index和Term，如果是空的话则返回-1，-1
func (r *RaftLog) LastIndexAndTerm() (term, index int64) {
	var lastIndex int64
	if lastIndex = r.Unstable.LastIndex(); lastIndex != -1 {
		lastIndex = r.Unstable.LastIndex()
	}
	if len(r.Unstable.Entries) == 0 {
		return -1, -1
	}
	lastTerm := r.Unstable.Entries[len(r.Unstable.Entries)-1].Term
	return lastIndex, lastTerm
}

// 更新传过来的committed的值
// 更新失败返回false，更新成功返回true
func (r *RaftLog) Commit(aimCommitted int64) bool {
	// 如果当前log的committed大于了传过来的committed
	// 即当前的提交领先，不用回滚，相等的时候可以返回true
	if aimCommitted < r.Committed {
		return false
	}
	if lastIndex, _ := r.LastIndexAndTerm(); lastIndex < aimCommitted {
		return false
	}
	// 将committed更新至最大值
	r.Committed = aimCommitted
	return false
}
