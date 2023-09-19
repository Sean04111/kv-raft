package kvraft

import (
	"crypto/rand"
	"kv-raft/labrpc"
	"math/big"
)

type Clerk struct {
	servers []*labrpc.ClientEnd
	// You will have to modify this struct.
	clientId  int64
	commandId int64
	leaderId  int64
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	// You'll have to add code here.
	ck.clientId = nrand()
	ck.commandId = 1 //这里注意起始位置是1，因为有dummy head log存在,在以后的debug中可能会有隐患！！！！
	ck.leaderId = 0
	return ck
}

// Get
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *Clerk) Get(key string) string {

	// You will have to modify this function.
	return ck.Command(&CommandArgs{
		Key: key,
		Ops: OpGet,
	})
}

// 这里不要加leaderid，在command里面填充
func (ck *Clerk) Put(key string, value string) {
	ck.Command(&CommandArgs{
		Key:   key,
		Value: value,
		Ops:   OpPut,
	})
	//ck.PutAppend(key, value, "Put")
}

//这里不要加leaderid，在command里面填充

func (ck *Clerk) Append(key string, value string) {
	//ck.PutAppend(key, value, "Append")
	ck.Command(&CommandArgs{
		Key:   key,
		Value: value,
		Ops:   OpAppend,
	})
}

// Command 这是一个发送的过程，如果出错，直接换service
func (ck *Clerk) Command(args *CommandArgs) string {
	args.CommandId, args.ClientId = ck.commandId, ck.clientId
	for {
		var reply CommandReply
		ok := ck.servers[ck.leaderId].Call("KVServer.Command", args, &reply)
		// fmt.Println(ck.clientId, "给", ck.leaderId, "发args,reply为", reply)
		if !ok || reply.Err == ErrWrongLeader || reply.Err == ErrTimeOut {
			ck.leaderId = (ck.leaderId + 1) % int64(len(ck.servers))
			continue
		} else {
			ck.commandId++
			return reply.Value
		}
	}
}
