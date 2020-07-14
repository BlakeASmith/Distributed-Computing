package raft

import "sync/atomic"
import "../labrpc"
import "time"
import "math/rand"



func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	rf.electionTimer = AsyncTimeout{
		duration: func () time.Duration { return time.Duration(150 + rand.Intn(300)) * time.Millisecond},
		action: rf.startElection,
	}

	rf.heartbeatTimer = AsyncTimeout{
		duration: func () time.Duration { return time.Duration(50) * time.Millisecond},
		action: rf.doHeartbeat,
		repeating: true,
	}

	rf.VotedFor = PersistedInt{
		value:LockedInt{value:-1},
		persistFunc: rf.persist,
	}

	rf.Term = PersistedInt{
		persistFunc: rf.persist,
	}

	rf.LeaderID = LockedInt{value:-1}

	rf.Log = Log{
		apply: applyCh,
		_log: make([]LogEntry, 0),
		rf: rf,
	}


	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	rf.electionTimer.set()

	return rf
}



func (rf *Raft) GetState() (int, bool) {
	return rf.Term.get(), rf.LeaderID.get() == rf.me
}

func (rf *Raft) Kill() {
	rf.log("KILLED")

	rf.electionTimer.cancel()
	rf.heartbeatTimer.cancel()
	rf.VotedFor.set(-1)
	rf.LeaderID.set(-1)
	rf.Term.set(0)
	rf.nextIndex = make([]LockedInt, len(rf.peers))
	rf.Log._log = make([]LogEntry, 0)
	rf.persist()
	atomic.StoreInt32(&rf.dead, 1)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}



func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

