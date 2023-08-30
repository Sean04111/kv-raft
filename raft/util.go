package raft

import "time"

// Debugging
const Debug = false

//系统日志的设计还需要进一步设计，这里实现了用es远程收集和zap本地磁盘文件写入的方法，
//es远程收集太慢了，会影响系统的正确性

func (rf *Raft) Record(eventtype string, event string) {
	if Debug {
		rf.logger.Info("  ", eventtype, "  ", event)
	}
}
func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
func (rf *Raft)LoggerRecorder(){
	for !rf.killed(){
		time.Sleep(rf.heartbeat)
		rf.logger.Info(rf.log.Print())
	}
}
