package kvraft

import (
	"kv-raft/common"
	"log"
	"time"
)

const ExecuteTimeout = 500 * time.Millisecond


const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type Optype uint8

const (
	OpGet Optype = iota
	OpPut
	OpAppend
)

func (op Optype) Opstring(optype Optype) string {
	switch op {
	case OpGet:
		return "get"
	case OpPut:
		return "put"
	case OpAppend:
		return "append"
	}
	return common.EmptyString
}

// CommandArgs
// 这里使用command封装所有的请求，是为了更好的操作commandId
type CommandArgs struct {
	Key       string
	Value     string
	ClientId  int64
	CommandId int64
	Ops       Optype
}
type CommandReply struct {
	Value string
	Err   common.Err
}

// OperationContext
// 为了保证幂等性，保留的命令提交的上下文
type OperationContext struct {
	LastCommandId int64
	LastReply     *CommandReply
}

// Command
// 为了发送而封装的cmd
type Command struct {
	*CommandArgs
}
