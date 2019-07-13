package raft

import (
	"errors"
	"sync"
)

type ValueMap struct {
	// TODO 加锁
	MapLock sync.Mutex
	Map     map[string]int64
}

func NewValueMap() *ValueMap {
	return &ValueMap{
		MapLock: sync.Mutex{},
		Map:     make(map[string]int64),
	}
}

func (v *ValueMap) AddValue(key string, value int64) error {
	if _, ok := v.Map[key]; ok {
		return errors.New("key has existed")
	}
	v.Map[key] = value
	return nil
}

func (v *ValueMap) UpdateValue(key string, value int64) error {
	if _, ok := v.Map[key]; !ok {
		return errors.New("key not exist")
	}
	v.Map[key] = value
	return nil
}

func (v *ValueMap) GetValue(key string) (value int64, err error) {
	value, ok := v.Map[key]
	if !ok {
		return 0, errors.New("key not exist")
	}
	return value, nil
}
