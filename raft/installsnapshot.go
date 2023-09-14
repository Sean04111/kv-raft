package raft

import (
	"bytes"
	"kv-raft/labgob"
	"strconv"
)

//加入2D后，需要注意日志的逻辑索引和物理索引
//这两者的关系应该是:
//logic index == physical index + rf.last_include_index

// InstallSnapshotArgs
// 这里的rpc消息实际上是leader和follower来交互更新快照的
// 具体参见paper section 7
type InstallSnapshotArgs struct {
	Term              int //leader's Term
	LeaderId          int //leader's id
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte //raw bytes of the snapshot chunk starting at offset
}
type InstallSnapshotReply struct {
	Term int //current Term
}

// InstallSnapshot

// Usually the snapshot will contain new information
//not already in the recipient’s log. In this case, the follower
//discards its entire log; it is all superseded by the snapshot
//and may possibly have uncommitted entries that conflict
//with the snapshot. If instead the follower receives a snapshot
//that describes a prefix of its log (due to retransmission or by mistake), then log entries covered by the snapshot are deleted but entries following the snapshot are still
//valid and must be retained.

func (rf *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) {
	rf.mu.Lock()
	reply.Term = rf.currentTerm

	//deal with the time-out message
	if args.Term < rf.currentTerm {
		rf.Record("leader_snapshot", "收到了leader的snapshot,但是我认为leader的term比我的小")
		rf.mu.Unlock()
		return
	}

	if args.LastIncludedIndex <= rf.lastincludeIndex {
		rf.mu.Unlock()
		return
	}

	//start trim the log
	//the length of the log can't be 0
	//so append a last_include_index entry
	//"dummy head"
	newlog := []Entry{{Term: args.LastIncludedTerm, Index: args.LastIncludedIndex, Cmd: -1}}

	//case 1 : the snapshot contain new information for the follower
	//the follower should trim all log and apply the snapshot

	if args.LastIncludedIndex > rf.LastIndex() {

		rf.log.Entries = newlog

	}
	//case 2 : the snapshot describes a prefix of the follower's log
	//logs covered by the snapshot are deleted,and the rest are retained
	if args.LastIncludedIndex <= rf.LastIndex() {
		for i := args.LastIncludedIndex + 1; i <= rf.LastIndex(); i++ {
			newlog = append(newlog, rf.EntryAt(i))
		}

		rf.log.Entries = newlog

	}
	if rf.commitIndex < args.LastIncludedIndex {
		rf.commitIndex = args.LastIncludedIndex
	}
	if rf.lastApplied <= args.LastIncludedIndex {
		rf.lastApplied = args.LastIncludedIndex
	}
	rf.lastincludeIndex = args.LastIncludedIndex
	rf.lastincludeTerm = args.LastIncludedTerm

	rf.Record("leader_snapshot", "收到的snapshot的index为"+strconv.Itoa(args.LastIncludedIndex)+"  现在我的log为"+rf.log.Print())

	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	e.Encode(rf.lastincludeIndex)
	e.Encode(rf.lastincludeTerm)
	data := w.Bytes()
	rf.persister.SaveStateAndSnapshot(data, args.Data)

	msg := ApplyMsg{
		SnapshotValid: true,
		Snapshot:      args.Data,
		SnapshotIndex: args.LastIncludedIndex,
		SnapshotTerm:  args.LastIncludedTerm,
	}
	rf.mu.Unlock()
	rf.applyCh <- msg
}

// leaderSendInstallSnapshotLocked
// leader给单个peer发送rpc消息并处理reply
// 占有锁的
func (rf *Raft) leaderSendInstallSnapshotLocked(server int, args *InstallSnapshotArgs) {

	reply := InstallSnapshotReply{}

	ok := rf.sendInstallSnapshot(server, args, &reply)

	if !ok {
		rf.Record("snapshot", "给 "+strconv.Itoa(server)+" 发送snapshot RPC超时")
		return
	}
	//处理过时RPC消息
	if rf.state != Leader || args.Term != rf.currentTerm {
		return
	}

	if reply.Term > rf.currentTerm {
		rf.MeetGreaterTerm(reply.Term)
		return
	}

	if args.Term == rf.currentTerm {
		rf.matchIndex[server] = max(args.LastIncludedIndex, rf.matchIndex[server])
		rf.nextIndex[server] = max(args.LastIncludedIndex+1, rf.nextIndex[server])
	}
}

// sendInstallSnapshot
// 调用lab rpc 发送消息
func (rf *Raft) sendInstallSnapshot(server int, args *InstallSnapshotArgs, reply *InstallSnapshotReply) bool {
	ok := rf.peers[server].Call("Raft.InstallSnapshot", args, reply)
	return ok
}
