package raft

import (
	"context"
	"os"
	"strconv"
	"time"
)

// DisplayLog
// 异步获取日志的实时状态
func (rf *Raft) DisplayLog(ctx context.Context) {
	writer, _ := os.Create("./logs/" + strconv.Itoa(rf.me) + "log.html")

	defer func(writer *os.File) {
		err := writer.Close()
		if err != nil {
			panic("fail to close the loglog file")
		}
	}(writer)
	for !rf.killed() {
		time.Sleep(1 * time.Millisecond)
		select {
		case <-ctx.Done():
			return
		default:
			_, err := writer.Write([]byte(rf.VirtualizeLog()))
			if err != nil {
				panic("file write error")
			}
		}
	}

}
func (rf *Raft) VirtualizeLog() string {
	logs := ""
	for i := 0; i < len(rf.log.Entries); i++ {
		cure := rf.log.Entries[i]
		command := ""
		if c, ok := cure.Cmd.(int); ok {
			command = strconv.Itoa(c)
		}
		if c, ok := cure.Cmd.(string); ok {
			command = c
		}
		log := "| " + strconv.Itoa(cure.Index) + " " + strconv.Itoa(cure.Term) + " " + command + " |"
		var curs string
		//这里先用绿色代表commit
		//红色为apply的
		if rf.commitIndex == rf.log.Entries[i].Index {
			log = "<span style = 'color:#09e509'>" + log + "</span>"
		}
		if rf.lastApplied == rf.log.Entries[i].Index {
			log = "<span style = 'background-color:rgba(100,100,100,0.5);'>" + log + "</span>"
		}
		curs = log
		logs += curs
	}
	return "<h4>" + logs + "</h4>"
}
