package raft

//这里是对log的设计
type Log struct{
	Entries []*Entry
}
type Entry struct{
	Term int
	Cmd interface{}
}
//获取最后一个日志索引
func (lg *Log)LastIndex()int{
	return len(lg.Entries)-1
}
//追加日志条目
func (lg *Log)Append(x Entry){
	lg.Entries = append(lg.Entries, &x)
}
//通过日志索引获取条目
func (lg *Log)EntryAt(index int)Entry{
	return *lg.Entries[index]
}