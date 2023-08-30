package raft

import "strconv"

//这里是对log的设计
type Log struct {
	Entries []Entry
}
type Entry struct {
	Term int
	Index int
	Cmd  interface{}
}

//获取最后一个有效的日志索引
func (lg *Log) LastIndex() int {
	return len(lg.Entries) - 1
}

//追加日志条目
func (lg *Log) Append(x Entry) {
	lg.Entries = append(lg.Entries, x)
}

//通过日志索引获取条目
func (lg *Log) EntryAt(index int) Entry {
		return lg.Entries[index]
}

//获取去区间内的entry,左闭右开,开的新slice
func (lg *Log) EntrySlice(from, end int) []Entry {
	ans := []Entry{}
	for i := from; i < end; i++ {
		ans = append(ans, lg.EntryAt(i))
	}
	return ans
}

//获取日志长度
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
		if value,ok:=v.Cmd.(string);ok{
			ans = ans + " index:" + strconv.Itoa(k) + " term:" + strconv.Itoa(v.Term) + " cmd:" + value + " | "
		}
		if value,ok:=v.Cmd.(int);ok{
			ans = ans + " index:" + strconv.Itoa(k) + " term:" + strconv.Itoa(v.Term) + " cmd:" + strconv.Itoa(value) + " | "
		}
	}
	return ans
}
