package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"math/rand"
	"strconv"
	"time"

	//	"bytes"
	"sync"
	"sync/atomic"

	//	"6.824/labgob"
	"kv-raft/labrpc"

	"go.uber.org/zap"
)

// ApplyMsg
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// A Go object implementing a single Raft peer.

type Pstate string

const Follower Pstate = "Follower"
const Leader Pstate = "Leader"
const Candidate Pstate = "Candidate"

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	applyCh      chan ApplyMsg //作为leader发给apply channel来应用
	applyCond    *sync.Cond    //用于唤醒apply channel 这里有nocopy使用引用
	state        Pstate        //节点的状态
	electionTime time.Time     //发起election的时间
	heartbeat    time.Duration //如果为leader的心跳时间
	// Your data here (2A, 2B, 2C).
	//持久性状态
	currentTerm int //这个节点最新的term
	votedFor    int //当前任期的获得选票的candidate的id
	log         Log //日志

	commitIndex int //这个节点已经提交的最大的index
	lastApplied int //这个节点apply到状态机的最大index(应该是和kv server有关的)

	nextIndex  []int //对于每个服务器，要发送给该服务器的下一个日志条目的索引（初始化为领导者的最后一个日志索引+1）
	matchIndex []int //对于每个服务器，已知在服务器上复制的最高日志条目的索引（初始化为0，单调增加）
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	logger *zap.SugaredLogger //日志写入器
}

// GetState return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	var term int
	var isleader bool
	// Your code here (2A).
	if rf.state == Leader {
		isleader = true
	}
	term = rf.currentTerm
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// CondInstallSnapshot
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// Snapshot
// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

// Start
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// Your code here (2B).
	if rf.state != Leader {
		return -1, rf.currentTerm, false
	} else {
		index = rf.log.LastIndex() + 1
		term = rf.currentTerm

		newentry := Entry{
			Term: term,
			Cmd:  command,
		}
		rf.log.Append(newentry)
		rf.logger.Info("leader接收到新的cmd: ", newentry, "此时leader的日志:", rf.log.Print())

		rf.LeaderAppendEntry(false)
		return index, term, true
	}
}

// Kill
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	rf.logger.Info("节点死亡")
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for rf.killed() == false {
		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().
		time.Sleep(rf.heartbeat)
		if rf.state == Leader {
			rf.mu.Lock()
			rf.setElectionTime()
			rf.LeaderAppendEntry(true)
			rf.mu.Unlock()
		}

		//这里直接比较now是否after就可以了因为follower在收到heartbeat的时候又会set一次electiontime
		if time.Now().After(rf.electionTime) {
			rf.mu.Lock()
			rf.setElectionTime()
			rf.LeaderElection()
			rf.mu.Unlock()
		}
	}
}

// 为该节点设置发起选举时间间隔(这个设置是以now为基准，提供发起election的时间)
func (rf *Raft) setElectionTime() {
	now := time.Now()
	due := rand.Intn(150)
	duration := time.Duration(150+due) * time.Millisecond
	rf.electionTime = now.Add(time.Duration(duration))
	rf.logger.Info("节点更新发起选举时间为" + strconv.Itoa(due) + "ms后")
}

// Applier 将日志提交到状态机(eg k-v server)
func (rf *Raft) applier() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	for !rf.killed() {
		if rf.commitIndex > rf.lastApplied && rf.lastApplied < rf.log.Len() {
			rf.lastApplied++
			applymgs := ApplyMsg{
				CommandValid: true,
				Command:      rf.log.EntryAt(rf.lastApplied).Cmd,
				CommandIndex: rf.lastApplied,
			}
			//这里可以直接发channel吗？
			rf.logger.Info("发送applymsg")
			rf.applyCh <- applymgs
			//

		} else {
			//applier等待
			rf.applyCond.Wait()
		}
	}
}

// 唤醒挂起的applier
func (rf *Raft) apply() {
	//唤醒applier
	rf.applyCond.Broadcast()
	rf.logger.Info("开启cond广播")
}

// Make
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).

	rf.LoggerInit()
	rf.logger.Info("节点创建")

	rf.dead = 0

	rf.applyCh = applyCh
	rf.state = Follower
	rf.setElectionTime()
	rf.heartbeat = 50 * time.Millisecond
	rf.currentTerm = 0
	rf.votedFor = -1 //这里需要初始化为-1
	rf.log = Log{Entries: []Entry{Entry{0, -1}}}
	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.applyCh = applyCh
	rf.applyCond = sync.NewCond(&rf.mu)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()
	//负责提交日志的
	go rf.applier()
	return rf

}

// peer遇到了term更高的节点的行为,
// 更新自己的term为新term,
// 对于candidate：失去竞选资格(在是否晋升逻辑哪里体现)
func (rf *Raft) MeetGreaterTerm(term int) {
	rf.currentTerm = term
	rf.state = Follower
	rf.votedFor = -1
	rf.logger.Info("节点遇到更高term,更新term为" + strconv.Itoa(term))
}
