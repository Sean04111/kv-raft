package raft

// RequestVoteArgs
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//

// 发送RequestVote消息的args，term为参选者自己的term,CandidateId为me,
// LastLogIndex为最后一个entry的索引,
// LastLogTerm为最后一个term
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int //candidate's term
	CandidateId  int //candidate
	LastLogIndex int //index of this candidate's last entry
	LastLogTerm  int //term of this candidate's last entry
}

// RequestVoteReply
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//

// RequVote的返回,
// VoteGranted代表是否获得选票
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int  //current term, for candidate to update itself
	VoteGranted bool //true means candidate got a vote
}

// RequestVote
// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintf("收到 %v 的投票邀请,term为 %v", args.CandidateId, args.Term)

	if args.Term > rf.currentTerm {
		rf.MeetGreaterTerm(args.Term)
	}
	if rf.state == Leader {
		DPrintf("节点 %v 已经是权威节点，故拒绝", rf.me)
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
	}
	if rf.state == Candidate {
		DPrintf("%v 的term不大于本节点,故拒绝放弃竞选", args.CandidateId)

		reply.Term = rf.currentTerm
		reply.VoteGranted = false
	}
	if rf.state == Follower {
		myindex := rf.log.LastIndex()
		myterm := rf.log.EntryAt(myindex).Term
		isuptodate := args.LastLogTerm > myterm || (args.LastLogTerm == myterm && myindex <= args.LastLogIndex)

		if args.Term < rf.currentTerm {
			reply.VoteGranted = false
			DPrintf("由于candidate %v 的term比本节点小,拒绝投票", args.CandidateId)
			return
		} else if (rf.votedFor == -1 || rf.votedFor ==args.CandidateId) && isuptodate {
			reply.VoteGranted = true
			rf.votedFor = args.CandidateId

			DPrintf("符合要求,投给 %v ", args.CandidateId)
			rf.persist()

			rf.setElectionTime()
		} else if rf.votedFor != -1 {
			DPrintf("本节点已经投票给 %v 故拒绝", rf.votedFor)
			reply.VoteGranted = false
		} else {
			DPrintf("不是uptodate,拒绝投票给 %v", args.CandidateId)
			reply.VoteGranted = false
		}
		reply.Term = rf.currentTerm
	}

}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//

// 向指定的server（在peers中的index）发送requestVote
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) RequestAndAdd(server int, args *RequestVoteArgs, vote *int) {
	reply := RequestVoteReply{}
	rf.sendRequestVote(server, args, &reply)

	if reply.Term > rf.currentTerm {
		DPrintf("发现更高的term : %v,放弃竞选", reply.Term)
		rf.MeetGreaterTerm(reply.Term)
	}
	if reply.VoteGranted && rf.state == Candidate {
		*vote++
		DPrintf("收到 %v 的投票", server)
		//这里判定是否达到多数派的做法是在每一次发送request请求的时候都判断一下
		if *vote > len(rf.peers)>>1 {
			//判断一下term是否被修改了
			//持有旧term的candidate是不能变成leader的
			if rf.currentTerm == args.Term {
				rf.UpGrade()
			}
		}
	}
}

// follower发起选举
func (rf *Raft) LeaderElection() {
	//首先state改为candidate
	//并且term++
	rf.state = Candidate
	rf.currentTerm++
	rf.votedFor = rf.me
	rf.persist()
	DPrintf("发起term为 %v 的选举", rf.currentTerm)

	args := &RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: rf.log.LastIndex(),
		LastLogTerm:  rf.log.EntryAt(rf.log.LastIndex()).Term,
	}
	//先给自己投一票
	vote := 1
	for k, _ := range rf.peers {
		if k != rf.me && rf.state == Candidate {
			go rf.RequestAndAdd(k, args, &vote)
		}
	}
}

func (rf *Raft) UpGrade() {
	if rf.state == Candidate {
		rf.state = Leader
	}
	DPrintf("节点 %v 成为leader", rf.me)

	//初始化nextindex和matchindex
	rf.entryIndexInitiallize()

	//发送ap消息
	rf.LeaderAppendEntry(true)

}

// 成为leader后初始化matchindex和nextindex
func (rf *Raft) entryIndexInitiallize() {
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))

	for k, _ := range rf.nextIndex {
		rf.nextIndex[k] = rf.log.LastIndex() + 1
	}
}
