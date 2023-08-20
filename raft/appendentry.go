package raft

import (
	"strconv"
)

// AppendEntryArgs
// append entry消息的定义
type AppendEntryArgs struct {
	Term         int     //leader 的term
	LeaderId     int     //leader在peers中的索引，这样follower就可以返回
	PrevLogIndex int     //新的日志项的前一个
	PrevLogTerm  int     //新的日志项的前一个的term
	Entries      []Entry //用于对齐log，作为心跳的时候为[]
	LeaderCommit int     //leader的commitIndex
}
type AppendEntryReply struct {
	Term     int  //follower的current term
	Success  bool //true：f包含prevLogIndex和prevLogTerm
	Conflict bool //follower和leader日志是否冲突
	Xindex   int  //follower匹配的index
	Xterm    int  //follower的匹配的term
}

// AppendEntry RPC
func (rf *Raft) AppendEntry(args *AppendEntryArgs, reply *AppendEntryReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.state == Candidate {
		if args.Term > rf.currentTerm {
			rf.logger.Info("收到新leader的心跳,放弃竞选")

			rf.MeetGreaterTerm(args.Term)
			reply.Term = rf.currentTerm
			reply.Success = true
			reply.Conflict = false
		} else {
			reply.Term = rf.currentTerm
			reply.Success = false
			reply.Conflict = false
		}
	}
	if rf.state == Leader {
		if args.Term > rf.currentTerm {
			rf.logger.Info("收到新leader" + strconv.Itoa(args.LeaderId) + "的心跳,leader下线")
			rf.MeetGreaterTerm(args.Term)
			reply.Term = rf.currentTerm
			reply.Success = true
			reply.Conflict = false
		} else {
			reply.Term = rf.currentTerm
			reply.Success = false
			reply.Conflict = false
		}
	}
	if rf.state == Follower {
		//这里接收到消息后就重新设置一下electiontime
		//防止重新发起election
		rf.setElectionTime()
		reply.Term = rf.currentTerm
		//由于leader断开导致的false

		if args.Term < rf.currentTerm {
			reply.Success = false
			reply.Conflict = false
		} else {
			rf.logger.Info("收到权威leader心跳,来自 "+strconv.Itoa(args.LeaderId), "args的entry为 ", args.Entries)
			followerlastindex := rf.log.LastIndex()

			//如果follower比leader日志落后
			if followerlastindex < args.PrevLogIndex {
				rf.logger.Info("发现log比leader落后")
				reply.Conflict = true
				reply.Success = false
				reply.Xindex = followerlastindex
				reply.Xterm = -1 //为了区分落后和不一致两种情况
				//这里直接结束了，交给下一轮
				return
			}
			//prelogindex不一致
			if rf.log.EntryAt(args.PrevLogIndex).Term != args.PrevLogTerm {
				rf.logger.Info("发现log和leader不一致")
				reply.Conflict = true
				reply.Success = false
				reply.Xindex = followerlastindex
				reply.Xterm = rf.log.EntryAt(args.PrevLogIndex).Term //为了区分落后和不一致两种情况
				return
			}
			//找到prelogindex,开始复制日志
			//这个做法有点丑，因为普通的心跳也会执行这一步
			//可能会产生bug
			//这里需要考虑一下entry里面到底是否需要带index ?
			//好像都走了这一步了？
			entry := append(rf.log.Entries[:args.PrevLogIndex+1], args.Entries...)
			rf.log.Entries = entry
			rf.logger.Info("目前日志为 ", rf.log.Print())
			reply.Conflict = false
			reply.Success = true
			reply.Xindex = rf.log.LastIndex()
			//处理提交的日志
			if args.LeaderCommit > rf.commitIndex {
				rf.commitIndex = min(args.LeaderCommit, rf.log.LastIndex())
				rf.apply()
			}
		}
	}
}

// 向每一个peer发送AP消息
func (rf *Raft) sendAppendEntry(server int, args *AppendEntryArgs, reply *AppendEntryReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntry", args, reply)

	return ok
}

// leader处理发给id为server的peer的AP消息的响应
func (rf *Raft) LeaderAP(server int, args *AppendEntryArgs) {
	reply := &AppendEntryReply{}
	ok := rf.sendAppendEntry(server, args, reply)
	if !ok {
		rf.logger.Info("给节点" + strconv.Itoa(server) + "发送心跳失败")
		return
	}
	rf.logger.Info("给节点" + strconv.Itoa(server) + "发送心跳成功")


	if reply.Term > rf.currentTerm {
		rf.logger.Info("遇到更大的term,leader下线")
		rf.MeetGreaterTerm(reply.Term)
		return
	}

	if reply.Success{
		//follower日志没有冲突且成功复制
		if !reply.Conflict {
			rf.logger.Info(server, " 日志没有冲突")
			if !reply.Conflict {
				//这里可能会有争议
				rf.nextIndex[server] = reply.Xindex + 1
				rf.matchIndex[server] = reply.Xindex
			}
			//不存在seccuss为true conflict为true的情况
			rf.LeaderTryCommit()
		}

	}else{
		//由于日志冲突导致false
		if reply.Conflict {
			//follower日志缺少
			if reply.Xterm == -1 {
				rf.logger.Info(server, " 的日志落后")
				rf.nextIndex[server] = reply.Xindex + 1
			} else {
				//follower 日志没对齐
				//这里采用递减,其实可以优化减少rpc协商次数
				rf.nextIndex[server]--
			}
		}
	}
	//是不是只有success才try?

}

// leader 发送AppendEntry消息,heatbeat看是否是心跳
// 如果heartbeat为true的话，那么不用leader有新entry也会发送
// 如果为false,那么只有leader有新entry才会发送
func (rf *Raft) LeaderAppendEntry(heartbeat bool) {

	leaderlastlogindex := rf.log.LastIndex()
	for k, _ := range rf.peers {
		if k == rf.me {
			rf.setElectionTime()
			continue
		}
		//heartbeat也需要承担日志对齐任务
		//如果对齐了：leaderlastindex = rf.nextIndex[k]-1
		//如果leader有新的：leaderlastindex >=rf.nextIndex[k]
		if leaderlastlogindex >= rf.nextIndex[k] || heartbeat {
			args := &AppendEntryArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me,
				LeaderCommit: rf.commitIndex,
			}
			//这里可能需要一个对nextindex的出界判断
			args.PrevLogIndex = rf.nextIndex[k] - 1
			args.PrevLogTerm = rf.log.EntryAt(args.PrevLogIndex).Term
			//切片截取的时候是可以左边界是可以取到切片长度的
			args.Entries = rf.log.EntrySlice(rf.nextIndex[k], rf.log.Len())
			go rf.LeaderAP(k, args)
		}
	}
}

// leader会尝试提交
// 提交的大概思路就是paper中figure2中的
// If there exists an N such that N > commitIndex, a majority
// of matchIndex[i] ≥ N, and log[N].term == currentTerm:
// set commitIndex = N
func (rf *Raft) LeaderTryCommit() {
	if rf.state != Leader {
		return
	}
	//N的取值范围在commitindex和len(log)之间

	for N := rf.commitIndex + 1; N < rf.log.Len(); N++ {
		if rf.log.EntryAt(N).Term != rf.currentTerm {
			continue
		}
		counter := 1
		for k, _ := range rf.peers {
			if rf.matchIndex[k] >= N {
				counter++
			}
		}
		//日志在多数派peer中得到了match
		if counter > len(rf.peers)/2 {
			rf.commitIndex = N
			rf.logger.Info("leader尝试发起commit ", N)
			rf.apply()
			break
		}
	}
}
