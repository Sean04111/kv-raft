package raft

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
// peer收到installSnapshot消息了做出的反应
func (rf *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) {
	rf.mu.Lock()

	reply.Term = rf.currentTerm
	if args.Term < rf.currentTerm {
		rf.mu.Unlock()
		return
	} else {
		rf.MeetGreaterTerm(args.Term)
	}

	reply.Term = rf.currentTerm

	if rf.lastincludeIndex >= args.LastIncludedIndex {
		return
	}

	//开始修剪follower的日志
	var newLog []Entry

	newLog = append(newLog, rf.EntryAt(args.LastIncludedIndex))

	for i := args.LastIncludedIndex + 1; i <= rf.LastIndex(); i++ {
		newLog = append(newLog, rf.EntryAt(i))
	}

	rf.lastincludeIndex = args.LastIncludedIndex
	rf.lastincludeTerm = args.LastIncludedTerm
	rf.log.Entries = newLog

	if args.LastIncludedIndex > rf.commitIndex {
		rf.commitIndex = args.LastIncludedIndex
	}
	if args.LastIncludedIndex > rf.lastApplied {
		rf.lastApplied = args.LastIncludedIndex
	}

	rf.persist()
	rf.persister.SaveRaftState(args.Data)

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
func (rf *Raft) leaderSendInstallSnapshotLocked(server int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	args := InstallSnapshotArgs{
		Term:              rf.currentTerm,
		LeaderId:          rf.me,
		LastIncludedIndex: rf.lastincludeIndex,
		LastIncludedTerm:  rf.lastincludeTerm,
		Data:              rf.persister.ReadSnapshot(),
	}
	reply := InstallSnapshotReply{}

	ok := rf.sendInstallSnapshot(server, &args, &reply)

	if ok {
		//处理过时RPC消息
		if rf.state != Leader || args.Term != rf.currentTerm {
			return
		}

		if reply.Term > rf.currentTerm {
			rf.MeetGreaterTerm(reply.Term)
			return
		}

		//这里为什么需要更新match index和next index?

		rf.matchIndex[server] = args.LastIncludedIndex
		rf.nextIndex[server] = args.LastIncludedIndex + 1

		return
	} else {
		rf.Record("snapshot", "leader send InstallSnapshot error (RPC)")
		return
	}
}

// sendInstallSnapshot
// 调用lab rpc 发送消息
func (rf *Raft) sendInstallSnapshot(server int, args *InstallSnapshotArgs, reply *InstallSnapshotReply) bool {
	ok := rf.peers[server].Call("Raft.InstallSnapshot", args, reply)
	return ok
}
