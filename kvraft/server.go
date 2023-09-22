package kvraft

import (
	"bytes"
	"kv-raft/common"
	"kv-raft/kvengine"
	"kv-raft/labgob"
	"kv-raft/labrpc"
	"kv-raft/raft"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Ops string
	Val string
}

type KVServer struct {
	mu      sync.RWMutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	dead    int32 // set by Kill()

	maxraftstate int // snapshot if log grows this big

	// Your definitions here.
	lastapplied int //用于记录该kv上一次apply的索引，用于过滤旧的信息
	kvDB        KVStateMachine
	lastOps     map[int64]OperationContext   //用于实现幂等性的hash表
	notifychan  map[int64]chan *CommandReply //leader回复给clerk的响应
}

// Kill
// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
	// Your code here, if desired.
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

// StartKVServer
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	//labgob.Register(Op{})
	labgob.Register(Command{})
	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate

	// You may need initialization code here.

	kv.dead = 0
	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)

	// You may need initialization code here.
	kv.lastapplied = 0
	kv.kvDB = kvengine.Start(kvengine.Config{
		DataDir:       "./datas-serivce" + strconv.Itoa(kv.me),
		Level0Size:    100,
		PartSize:      4,
		Threshold:     30,
		CheckInterval: 3,
	})
	// kv.kvDB = NewMKV()
	kv.lastOps = map[int64]OperationContext{}
	kv.notifychan = map[int64]chan *CommandReply{}
	kv.RestoreSnapshot(persister.ReadSnapshot())
	go kv.applier()

	return kv
}
func (kv *KVServer) applier() {
	for !kv.killed() {
		//不断从applych中拿
		select {
		case message := <-kv.applyCh:

			if message.CommandValid {
				kv.mu.Lock()
				//如果拿到的信息是旧的，直接跳过
				if message.CommandIndex <= kv.lastapplied {
					kv.mu.Unlock()
					continue
				}

				kv.lastapplied = message.CommandIndex

				command := message.Command.(Command)
				//准备在DB操作
				reply := &CommandReply{}

				//幂等性判断
				if command.Ops != OpGet && kv.CheckExed(command.CommandArgs) {
					reply = kv.lastOps[command.ClientId].LastReply
				} else {
					//新的command,直接在DB中查询
					reply.Value, reply.Err = kv.ApplytoStartMachine(command)

					//幂等性记录
					if command.Ops != OpGet {
						kv.lastOps[command.ClientId] = OperationContext{
							LastCommandId: command.CommandId,
							LastReply:     reply,
						}
					}
				}

				//还需要判断一下这个此时raft的状态是否改变了
				if curterm, isLeader := kv.rf.GetState(); isLeader && curterm == message.CommandTerm {
					ch := kv.GetNotifyChan(int64(message.CommandIndex))
					ch <- reply
				}

				//判断是否需要执行快照
				if kv.NeedSnapShot() {
					kv.TakeSnapShot(message.CommandIndex)
				}
				kv.mu.Unlock()
			} else if message.SnapshotValid {
				//如果applier接收到了leader要求打快照的信息
				kv.mu.Lock()
				kv.RestoreSnapshot(message.Snapshot)
				kv.lastapplied = message.SnapshotIndex
				kv.mu.Unlock()
			}
		}
	}
}

// 当applier收到一个commandValid的时候,说明raft状态得到了更新
// 判断一下是否需要打快照
func (kv *KVServer) NeedSnapShot() bool {
	return kv.maxraftstate != -1 && kv.rf.Persister.RaftStateSize() >= kv.maxraftstate
}

// kv server下的rf主动打快照
func (kv *KVServer) TakeSnapShot(index int) {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(kv.kvDB)
	e.Encode(kv.lastOps)
	kv.rf.Snapshot(index, w.Bytes())
}

// kv 安装快照
func (kv *KVServer) RestoreSnapshot(snapshot []byte) {
	if snapshot == nil || len(snapshot) == 0 {
		return
	}
	r := bytes.NewBuffer(snapshot)
	d := labgob.NewDecoder(r)
	var NewDB MKV
	var NewlastOps map[int64]OperationContext
	if d.Decode(&NewDB) != nil || d.Decode(&NewlastOps) != nil {
		panic("decode error !")
	} else {
		kv.kvDB = &NewDB
		kv.lastOps = NewlastOps
	}

}

// GetNotifyChan
// 封装的一个判断当前commandid对应的notifychan是否有chan
// 如果没有建一个，如果有直接返回
// 读写一个nil的channel都会阻塞
func (kv *KVServer) GetNotifyChan(index int64) chan *CommandReply {
	if _, ok := kv.notifychan[index]; !ok {
		kv.notifychan[index] = make(chan *CommandReply, 1)
	}
	return kv.notifychan[index]
}

// ApplytoStartMachine service 在DB上读写操作
func (kv *KVServer) ApplytoStartMachine(cmd Command) (string, common.Err) {
	switch cmd.Ops {
	case OpGet:
		return kv.kvDB.Get(cmd.Key)
	case OpPut:
		err := kv.kvDB.Put(cmd.Key, cmd.Value)
		return common.EmptyString, err
	case OpAppend:
		err := kv.kvDB.Append(cmd.Key, cmd.Value)
		return common.EmptyString, err
	}
	return common.EmptyString, common.EmptyString
}

// Command handler
// command handler需要实现当底层raft没有把这条日志放到applychan之前，需要阻塞给clerk的回复
// applier一直监听该kv的applychan
func (kv *KVServer) Command(args *CommandArgs, reply *CommandReply) {
	// fmt.Println("收到", args.ClientId, "的args,opstype:", args.Ops)
	if kv.killed() {
		return
	}
	kv.mu.RLock()

	//幂等性检查
	//如果这个clerk的操作已经被执行过了
	if kv.CheckExed(args) && args.Ops != OpGet {
		// fmt.Println("a")
		lastreply := kv.lastOps[args.ClientId].LastReply
		reply.Value = lastreply.Value
		reply.Err = lastreply.Err
		kv.mu.RUnlock()
		return
	}

	//是新的请求,没有执行过
	kv.mu.RUnlock()
	cmd := Command{args}

	index, _, ok := kv.rf.Start(cmd)
	// fmt.Println(args.ClientId, "的这次index为", index)
	//如果当前service不是raft leader的service
	if !ok {
		reply.Err = common.ErrWrongLeader
		reply.Value = common.EmptyString
		return
	}

	kv.mu.Lock()
	ch := kv.GetNotifyChan(int64(index))
	kv.mu.Unlock()
	select {
	case msg := <-ch:
		reply.Value = msg.Value
		reply.Err = msg.Err
	case <-time.After(ExecuteTimeout):
		reply.Err = common.ErrTimeOut
		reply.Value = common.EmptyString
	}
	go func() {
		kv.mu.Lock()
		delete(kv.notifychan, int64(index))
		kv.mu.Unlock()
	}()
}

// CheckExed 如果这个args在之前已经被操作了，那么返回true
// 如果没有没执行，返回false
func (kv *KVServer) CheckExed(args *CommandArgs) bool {
	value, ok := kv.lastOps[args.ClientId]
	return ok && args.CommandId <= value.LastCommandId
}
