package raft

type UnstableLog struct {
	Snapshot *Snapshot
	Entries  []Entry
}

func (u *UnstableLog) FirstIndex() int64 {
	if u.Snapshot != nil {
		return u.Snapshot.Index
	}
	if len(u.Entries) != 0 {
		return u.Entries[0].Index
	}
	return 0
}

func (u *UnstableLog) LastIndex() int64 {
	if len(u.Entries) != 0 {
		return u.Entries[len(u.Entries)-1].Index
	}
	if u.Snapshot != nil {
		return u.Snapshot.Index
	}
	return 0
}

//
func (u *UnstableLog) Term(index int64) int64 {
	for _, entry := range u.Entries {
		if entry.Index == index {
			return entry.Term
		}
	}
	if u.Snapshot != nil && u.Snapshot.Index == index {
		return u.Snapshot.Term
	}
	return 0
}

func (u *UnstableLog) AppendEntry(entry ...Entry) {
	u.Entries = append(u.Entries, entry...)
}

// Snapshot中的Data包括了全部的entry？
// TODO：判断是否能够添加成功
func (u *UnstableLog) AppendSnapshot(snapshot Snapshot) {
	u.Snapshot = &snapshot
	u.ShrinkEntry(snapshot.Index)
}

// 在添加Snapshot后，将Snapshot index之前的entry数组全部去掉
func (u *UnstableLog) ShrinkEntry(index int64) bool {
	pos := -1
	for i, e := range u.Entries {
		if e.Index == index {
			pos = i
		}
	}
	if pos != -1 {
		newEntry := make([]Entry, len(u.Entries)-pos, len(u.Entries)-pos)
		copy(newEntry, u.Entries[pos:len(u.Entries)])
		u.Entries = newEntry
		return true
	}
	return false
}
