package raft

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

const (
	TestLogFilePrefix = "F:\\_personal_github\\Raft_Demo\\src\\"
	TestLogFileSuffix = ".log"
)

// 持久化的LOG
type StableLog struct {
	FileName string
}

func (s *StableLog) FirstIndex() int64 {
	file := s.OpenReadFile()
	if file != nil {
		reader := bufio.NewReader(file)
		line, _, err := reader.ReadLine()
		if err != nil {
			return -1
		}
		var entry Entry
		json.Unmarshal(line, &entry)
		if entry.Data == nil {
			return -1
		}
		return entry.Index
	}
	return -1
}
func (s *StableLog) LastIndex() int64 {
	file := s.OpenReadFile()
	if file != nil {
		reader := bufio.NewReader(file)
		var bytes []byte
		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				break
			}
			bytes = line
		}
		var entry Entry
		json.Unmarshal(bytes, &entry)
		if entry.Data == nil {
			return -1
		}
		return entry.Index
	}
	return -1
}
func (s *StableLog) Term(index int64) int64 {
	// TODO 暂时返回 -1 应该遍历然后查找
	return -1
}

func (s *StableLog) Write(entry Entry) {
	file, e := os.OpenFile(s.FileName, os.O_RDONLY|os.O_APPEND, 0777)
	if e != nil {
		fmt.Println("openFile error")
	}
	defer file.Close()
	bytes, _ := json.Marshal(entry)
	file.Write(bytes)
	file.Write([]byte("\n"))
}

func (s *StableLog) OpenReadFile() *os.File {
	file, e := os.OpenFile(s.FileName, os.O_RDONLY, 0777)
	if e != nil {
		fmt.Println("openFile error")
		return nil
	}
	return file
}
