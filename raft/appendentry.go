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
	reply.Success = false
	if rf.state == Candidate {
		if args.Term > rf.currentTerm {
			event := "作为candidate收到来自" + strconv.Itoa(args.LeaderId) + "的心跳,对方term更高,放弃竞选"
			rf.Record("心跳", event)
			rf.MeetGreaterTerm(args.Term)
			reply.Term = rf.currentTerm
		} else {
			reply.Term = rf.currentTerm
		}
		return
	}
	if rf.state == Leader {
		if args.Term > rf.currentTerm {
			event := "收到新leader" + strconv.Itoa(args.LeaderId) + "的心跳,leader下线"
			rf.Record("心跳", event)
			rf.MeetGreaterTerm(args.Term)
			reply.Term = rf.currentTerm
		} else {
			reply.Term = rf.currentTerm
		}
		return
	}
	if rf.state == Follower {
		//这里接收到消息后就重新设置一下electiontime
		//防止重新发起election
		reply.Term = rf.currentTerm
		//由于leader断开导致的false

		if args.Term < rf.currentTerm {

			return
		}
		if args.Term > rf.currentTerm {
			rf.MeetGreaterTerm(args.Term)

			return
		}

		if args.Term == rf.currentTerm {
			rf.setElectionTimeUnlocked()

			rf.Record("心跳", "收到权威leader心跳,来自 "+strconv.Itoa(args.LeaderId))

			followerlastindex := rf.LastIndex()

			//如果自身最后的快照比prev小说明中间有日志确实
			//如果follower比leader日志落后
			if followerlastindex < args.PrevLogIndex || rf.lastincludeIndex > args.PrevLogIndex {
				rf.Record("日志对齐", "发现log比leader落后")
				reply.Conflict = true
				reply.Xindex = followerlastindex + 1
				reply.Xterm = -1 //为了区分落后和不一致两种情况
				//这里直接结束了，交给下一轮
				return
			}
			//lastincludedIndex永远在是日志中的最后一项
			//prelogindex不一致
			if rf.EntryAt(args.PrevLogIndex).Term != args.PrevLogTerm {
				rf.Record("日志对齐", "发现log和leader不一致")
				reply.Conflict = true

				reply.Xterm = rf.EntryAt(args.PrevLogIndex).Term //为了区分落后和不一致两种情况
				//这里xindex是返回Xterm中的第一个索引
				//If a follower does have prevLogIndex in its log, but the term does not match, it should return conflictTerm = log[prevLogIndex].Term, and then search its log for the first index whose entry has term equal to conflictTerm.
				//来自guide-book
				for xindex := args.PrevLogIndex; xindex > rf.lastincludeIndex; xindex-- {
					if rf.EntryAt(xindex-1).Term != reply.Xterm {
						reply.Xindex = xindex
						break
					}
				}
				return
			}
			//找到prelogindex,开始复制日志
			//这里entry是批量复制的
			//这里需求是tun掉冲突entry的后面的所有的entry
			for inx, entry := range args.Entries {
				if entry.Index <= rf.LastIndex() && entry.Term != rf.EntryAt(entry.Index).Term {
					rf.log.Entries = rf.log.Entries[:entry.Index]
					rf.persist()
				}
				if entry.Index > rf.LastIndex() {
					rf.log.Entries = append(rf.log.Entries, args.Entries[inx:]...)
					rf.persist()
					break
				}
			}
			rf.Record("日志对齐", "本节点日志和leader已经对齐,本节点日志为"+rf.log.Print())

			/*
				entry := append(rf.log.Entries[:args.PrevLogIndex+1], args.Entries...)
				rf.log.Entries = entry
			*/
			reply.Success = true

			//处理提交的日志
			if args.LeaderCommit > rf.commitIndex {
				rf.commitIndex = min(args.LeaderCommit, rf.LastIndex())
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

// LeaderAPUnlocked
// leader处理发给id为server的peer的AP消息的响应
func (rf *Raft) LeaderAPUnlocked(server int, args *AppendEntryArgs) {
	if rf.state != Leader {
		return
	}

	reply := &AppendEntryReply{}
	ok := rf.sendAppendEntry(server, args, reply)
	if !ok {
		return
	}
	rf.Record("心跳", "给节点 "+strconv.Itoa(server)+" 发送心跳成功")
	if reply.Term > rf.currentTerm {
		rf.Record("心跳", "遇到更大的term,leader下线")
		rf.MeetGreaterTerm(reply.Term)
		return
	}
	//这里是处理过时的rpc消息
	if rf.currentTerm == args.Term {
		if reply.Success {
			//follower日志没有冲突且成功复制
			rf.Record("日志对齐", strconv.Itoa(server)+" 日志没有冲突")
			if !reply.Conflict {
				match := args.PrevLogIndex + len(args.Entries)
				next := match + 1
				rf.nextIndex[server] = next
				rf.matchIndex[server] = match
			}
			//不存在seccuss为true conflict为true的情况

		} else if reply.Conflict {
			//由于日志冲突导致false
			//follower日志缺少
			if reply.Xterm == -1 {
				rf.Record("日志对齐", strconv.Itoa(server)+"的日志落后")
				rf.nextIndex[server] = reply.Xindex
			} else {
				//follower 日志没对齐
				//prelog不一致
				//这里采用搜寻做法可以优化减少rpc协商次数

				newindex := rf.FindLastIndexofReplyTerm(reply.Xterm)
				if newindex > 0 {
					rf.nextIndex[server] = newindex + 1
				} else {
					rf.nextIndex[server] = reply.Xindex
				}

			}
		}

		rf.LeaderTryCommitUnlocked()

	} else {
		return
	}

	//是不是只有success才try?

}

// FindLastIndexofReplyTerm
// 为了减少日志同步的时候的协商次数
// 如果日志prelog和follower不一致，那么用此方法
// 找出在leader中最后一个一致的index
// Upon receiving a conflict response, the leader should first search its log for conflictTerm. If it finds an entry in its log with that term, it should set nextIndex to be the one beyond the index of the last entry in that term in its log.If it does not find an entry with that term, it should set nextIndex = conflictIndex.
func (rf *Raft) FindLastIndexofReplyTerm(replyterm int) int {
	for i := rf.LastIndex(); i > rf.lastincludeIndex; i-- {
		if rf.EntryAt(i).Term == replyterm {
			return i
		} else if rf.EntryAt(i).Term < replyterm {
			break
		}
	}
	return -1
}

// LeaderAppendEntryLocked
// leader 发送AppendEntry消息,heatbeat看是否是心跳
// 如果heartbeat为true的话，那么不用leader有新entry也会发送
// 如果为false,那么只有leader有新entry才会发送
// 需要占有锁
func (rf *Raft) LeaderAppendEntryLocked(heartbeat bool) {
	leaderlastlogindex := rf.LastIndex()
	for k, _ := range rf.peers {
		if k == rf.me {
			rf.setElectionTimeUnlocked()
			continue
		}
		if rf.nextIndex[k]-1 < rf.lastincludeIndex {
			//这里可以看出其实整个项目的锁结构是非常混乱的,而且锁的粒度还不小
			//这里必须要优化一下
			go rf.leaderSendInstallSnapshotLocked(k)
			return
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
			//这里需要一个对nextindex的出界判断
			nextindex := rf.nextIndex[k]
			if nextindex <= rf.lastincludeIndex {
				nextindex = rf.lastincludeIndex + 1
			}
			if nextindex > rf.log.Len() {
				nextindex = leaderlastlogindex
			}
			args.PrevLogIndex = nextindex - 1
			args.PrevLogTerm = rf.EntryAt(args.PrevLogIndex).Term
			//切片截取的时候是可以左边界是可以取到切片长度的
			//这里其实就实现了no-op日志复制解决了幽灵日志
			//初始化的时候nextindex会被初始化为len
			//那么此时args的entry就是空的,也就是no-op
			args.Entries = append([]Entry(nil), rf.log.Entries[nextindex-rf.lastincludeIndex:]...)
			go rf.LeaderAPUnlocked(k, args)
		}
	}
}

// LeaderTryCommitUnlocked
// leader会尝试提交
// 提交的大概思路就是paper中figure2中的
// If there exists an N such that N > commitIndex, a majority
// of matchIndex[i] ≥ N, and log[N].term == currentTerm:
// set commitIndex = N
func (rf *Raft) LeaderTryCommitUnlocked() {
	if rf.state != Leader {
		return
	}
	//N的取值范围在commitindex和len(log)之间

	for N := rf.lastincludeIndex + 1; N < rf.log.Len(); N++ {
		if rf.EntryAt(N).Term != rf.currentTerm {
			continue
		}
		counter := 1
		for k := range rf.peers {
			if rf.matchIndex[k] >= N && k != rf.me {
				counter++
			}
			//日志在多数派peer中得到了match
			//leader提交条件
			if counter > len(rf.peers)/2 {
				rf.commitIndex = N
				rf.Record("日志提交", "leader尝试发起commitindex="+strconv.Itoa(N))
				rf.apply()
				break
			}
		}

	}
}
