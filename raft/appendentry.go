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

func (rf *Raft) AppendEntry(args *AppendEntryArgs, reply *AppendEntryReply) {
	rf.mu.Lock()
	if rf.state==Candidate{
		if args.Term>rf.currentTerm{
			rf.logger.Info("收到新leader的心跳,放弃竞选")
			rf.MeetGreaterTerm(args.Term)
			reply.Term = rf.currentTerm
			reply.Success = true
		}else{
			reply.Term = rf.currentTerm
			reply.Success = true
		}
	}
	if rf.state ==Leader{
		if args.Term>rf.currentTerm{
			rf.logger.Info("收到新leader的心跳,leader下线")
			rf.MeetGreaterTerm(args.Term)
			reply.Term = rf.currentTerm
			reply.Success = true
		}else{
			reply.Term = rf.currentTerm
			reply.Success = true
		}
	}
	if rf.state ==Follower{
		//这里接收到消息后就重新设置一下electiontime
		//防止重新发起election
		rf.logger.Info("收到心跳,来自",args.LeaderId)
		rf.setElectionTime()
		reply.Term = rf.currentTerm
		reply.Success = true
	}
	rf.mu.Unlock()




}
func (rf *Raft) sendAppendEntry(server int, args *AppendEntryArgs ,reply *AppendEntryReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntry", args, reply)
	return ok
}
func (rf *Raft)LeaderAP(server int,args *AppendEntryArgs){
	reply:= &AppendEntryReply{}
	rf.sendAppendEntry(server,args,reply)
	if reply.Term>rf.currentTerm{
		rf.logger.Info("遇到更大的term,leader下线")
		rf.MeetGreaterTerm(reply.Term)
	}
	
}
//leader 发送AppendEntry消息
func (rf *Raft)LeaderAppendEntry(){

	args:=&AppendEntryArgs{
		Term:rf.currentTerm,
		LeaderId: rf.me,
	}
	for k,_:=range rf.peers{
		if k!=rf.me &&rf.state==Leader{
			go rf.LeaderAP(k,args)
		}
	}
}