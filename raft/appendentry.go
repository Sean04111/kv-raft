package raft

// AppendEntryArgs
//append entry消息的定义
type AppendEntryArgs struct {
	Term         int        //leader 的term
	LeaderId     int        //leader在peers中的索引，这样follower就可以返回
	PrevLogIndex int        //新的日志项的前一个
	PrevLogTerm  int        //新的日志项的前一个的term
	Entries      []Entry //用于对齐log，作为心跳的时候为[]
	LeaderCommit int        //leader的commitIndex
}
type AppendEntryReply struct {
	Term    int  //follower的current term
	Success bool //true：f包含prevLogIndex和prevLogTerm
}

func (rf *Raft) AppendEntry(args *AppendEntryArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.state==Candidate{
		if args.Term>rf.currentTerm{
			rf.MeetGreaterTerm(args.Term)
		}
	}
	//这里接收到消息后就重新设置一下electiontime
	//防止重新发起election
	rf.setElectionTime()




}
func (rf *Raft) sendAppendEntry(server int, args *AppendEntryArgs ,reply *AppendEntryReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntry", args, reply)
	return ok
}
//leader 发送AppendEntry消息
func (rf *Raft)LeaderAppendEntry(){

	args:=&AppendEntryArgs{
		Term:rf.currentTerm,
		LeaderId: rf.me,
	}
	for k,_:=range rf.peers{
		if k!=rf.me{
			go rf.sendAppendEntry(k,args,&AppendEntryReply{})
		}
	}
}