package raft

// 持久化的LOG
type StableLog struct {
}

func (s *StableLog) FirstIndex() int64 {
	return 0
}
func (s *StableLog) LastIndex() int64 {
	return 0
}
func (s *StableLog) Term(index int64) int64 {
	return 0
}
