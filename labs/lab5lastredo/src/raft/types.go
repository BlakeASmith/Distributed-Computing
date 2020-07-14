package raft

import "../labrpc"
import "sync"
import "log"
import "fmt"

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	Term PersistedInt

	VotedFor PersistedInt
	LeaderID LockedInt

	electionTimer AsyncTimeout
	heartbeatTimer AsyncTimeout

	Log Log

	nextIndex []LockedInt
}

type LogEntry struct {
	Term int
	Command interface{}
	Index int
}

type Log struct {
	_log []LogEntry
	commitInd LockedInt
	apply chan ApplyMsg
	rf *Raft
	sizemu sync.Mutex
}

func (lg *Log) Size() int {
	lg.sizemu.Lock()
	defer lg.sizemu.Unlock()
	return len(lg._log)
}

func (lg *Log) getCommitInd() int {
	return lg.commitInd.get()+1
}

func (lg *Log) add(it LogEntry) *Log{
	if it.Index == lg.Size(){
		lg._log = append(lg._log, it)
	} else if it.Index < lg.Size() {
		lg._log[it.Index] = it
		lg._log = lg._log[:it.Index+1]
	}
	lg.rf.persist()
	return lg
}

func (lg *Log) extend(lst []LogEntry) *Log{
	for _, it := range lst {
		lg.add(it)
	}
	return lg
}

func (lg *Log) commitTo(ind int, id int){
	//assert(lg.commitInd.get() <= ind, 
	//fmt.Sprintf("Commit index too big: have %d, commiting up to %d at %d", 
	//lg.commitInd.get(), ind, id))
	if lg.commitInd.get() > ind{
		fmt.Printf("Commit index too big: have %d, commiting up to %d at %d\n",
			lg.commitInd.get(), ind, id)
		lg.commitInd.set(0)
	}
	for _, it := range lg._log[lg.commitInd.get():ind] {
		//assert(lg.commitInd.get() < lg.Size(), "")
		lg.commitInd.inc()
		msg := ApplyMsg{
			CommandValid: true,
			Command: it.Command,
			CommandIndex: lg.commitInd.get(),
		}
		log.Println("applied ", msg, " at", id)
		lg.apply <- msg
	}

	lg.rf.persist()
}

func (lg *Log) PrevLogTerm() int {
	if lg.Size() == 0 {return -1}
	return lg._log[lg.Size()-1].Term
}

type RequestVoteArgs struct {
	LeaderID int
	ID int
	Term int
	LastLogTerm int
	LogLen int
	CommitInd int
}

type RequestVoteReply struct {
	Vote bool
	Term int
}

type AppendEntryArgs struct {
	Term int
	LeaderID int
	PrevLogIndex int
	PrevLogTerm int
	Entries []LogEntry
	LeaderCommitIndex int
	LogLen int
}

type AppendEntryReply struct {
	Term int
	Success bool
	ID int
}


type LockedInt struct {
	value int
	mu sync.Mutex
}

func (lc *LockedInt) inc() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.value += 1
	return lc.value
}

func (lc *LockedInt) dec() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.value -= 1
	return lc.value
}

func (lc *LockedInt) plus(a int) int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.value += a
	return lc.value
}

func (lc *LockedInt) get() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.value
}

func (lc *LockedInt) set(value int) int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.value = value
	return lc.value
}

type PersistedInt struct {
	value LockedInt
	persistFunc func()
}


func (pi *PersistedInt) inc() int {
	pi.value.inc()
	pi.persistFunc()
	return pi.value.get()
}

func (pi *PersistedInt) dec() int {
	pi.value.inc()
	pi.persistFunc()
	return pi.value.get()
}

func (pi *PersistedInt) get() int {
	return pi.value.get()
}

func (pi *PersistedInt) set(value int) int {
	pi.value.set(value)
	pi.persistFunc()
	return pi.value.get()
}



