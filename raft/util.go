package raft

import "log"

// Debugging
const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}
func max(a ,b int)int{
	if a>b{
		return a
	}else{
		return b
	}
}
