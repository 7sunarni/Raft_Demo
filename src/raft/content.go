package raft

import (
	"errors"
	"sync"
)

type ValueMap struct {
	m   sync.Mutex
	Map map[string]int64
}

func NewValueMap() *ValueMap {
	return &ValueMap{
		m:   sync.Mutex{},
		Map: make(map[string]int64),
	}
}

func (v *ValueMap) AddValue(key string, value int64) error {
	v.m.Lock()
	defer v.m.Unlock()
	if _, ok := v.Map[key]; ok {
		return errors.New("key has existed")
	}
	v.Map[key] = value
	return nil
}

func (v *ValueMap) UpdateValue(key string, value int64) error {
	v.m.Lock()
	defer v.m.Unlock()
	if _, ok := v.Map[key]; !ok {
		return errors.New("key not exist")
	}
	v.Map[key] = value
	return nil
}

func (v *ValueMap) GetValue(key string) (value int64, err error) {
	v.m.Lock()
	defer v.m.Unlock()
	value, ok := v.Map[key]
	if !ok {
		return 0, errors.New("key not exist")
	}
	return value, nil
}

func (v *ValueMap) DeleteValue(key string) (err error) {
	v.m.Lock()
	defer v.m.Unlock()
	_, ok := v.Map[key]
	if !ok {
		return errors.New("key not exist")
	}
	delete(v.Map, key)
	return nil
}
