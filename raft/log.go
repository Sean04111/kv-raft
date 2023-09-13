package raft

import (
	"strconv"
)

// Log
// 这里是对log的设计
type Log struct {
	Entries []Entry
}

type Entry struct {
	Term  int
	Index int
	Cmd   interface{}
}

// LastIndex
// 获取最后一个合法的日志索引
// logic index
func (rf *Raft) LastIndex() int {
	return len(rf.log.Entries) - 1 + rf.lastincludeIndex
}

// Append
// 追加日志条目
func (rf *Raft) Append(x Entry) {
	rf.log.Entries = append(rf.log.Entries, x)
}

// EntryAt
// 通过日志索引获取条目
// 这里根据2D做了修改
func (rf *Raft) EntryAt(index int) Entry {
	if index > rf.LastIndex() {
		return rf.log.Entries[0]
	}
	newindex := index - rf.lastincludeIndex
	if newindex <= 0 {
		return rf.log.Entries[0]
	}
	return rf.log.Entries[newindex]
}

// Len
// 获取日志长度
func (lg *Log) Len() int {
	return len(lg.Entries)
}
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (lg *Log) Print() string {
	ans := "| "
	for k, v := range lg.Entries {
		if value, ok := v.Cmd.(string); ok {
			ans = ans + " index:" + strconv.Itoa(lg.Entries[k].Index) + " term:" + strconv.Itoa(v.Term) + " cmd:" + value + " | "
		}
		if value, ok := v.Cmd.(int); ok {
			ans = ans + " index:" + strconv.Itoa(lg.Entries[k].Index) + " term:" + strconv.Itoa(v.Term) + " cmd:" + strconv.Itoa(value) + " | "
		}
	}
	return ans
}
