package raft

import "testing"

// 基本的增加删除测试
func TestContentMap(t *testing.T) {
	m := NewValueMap()
	{
		m.AddValue("APDO", 12)
		value, err := m.GetValue("APDO")
		if err != nil || value != 12 {
			t.Fatal("Add error", value, err)
		}
	}
	{
		m.UpdateValue("APDO", 24)
		value, err := m.GetValue("APDO")
		if err != nil || value != 24 {
			t.Fatal("Update error", value, err)
		}
	}
	{
		m.DeleteValue("APDO")
		value, err := m.GetValue("APDO")
		if err == nil {
			t.Fatal("Delete error", value, err)
		}
	}
}

// 并发测试
func TestWithGoroutine(t *testing.T) {

}
