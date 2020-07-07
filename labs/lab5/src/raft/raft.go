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

import "sync"
import "sync/atomic"
import "../labrpc"
import "time"
import "log"
import "math/rand"

// import "bytes"
// import "../labgob"


//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()
	isLeader  bool
	term      int
	applychan chan ApplyMsg
	hasvoted  bool

	// used to inform a timer goroutine that the timer was cancelled
	electionID int
	nelectiontimeout int
	heartbeatID int
	nheartbeattimeout int
}

type AppendEntry struct {
	Term int

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	log.Println("state requested from ", rf.me, " isLeader: ", rf.isLeader, "term: ", rf.term)
	return rf.term, rf.isLeader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
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


//
// restore previously persisted state.
//
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

func (rf *Raft) revokeLeadership() {
	log.Println("revoking leadership from ", rf.me)
	rf.mu.Lock()
	if !rf.isLeader {
		rf.mu.Unlock()
		return
	}
	if rf.isLeader {
		rf.heartbeatID++
	}
	rf.isLeader = false
	rf.mu.Unlock()
	go rf.startElectionTimeout()
}

// Start a new timeout (this is blocking) and perform the given action if the 
// timeout is not cancelled
func (rf *Raft) setTimeout(id int, currentID *int, counter *int,  length time.Duration, action func ()){
	time.Sleep(length)
	log.Println("current ID for timer ", *currentID, " server ", rf.me)
	if id == *currentID{
		log.Println("performing action with id ", id, " at ", rf.me)
		action()
	}
	*counter--
}

func (rf *Raft) startElectionTimeout(){
	log.Println("setting election timeout at ", rf.me)
	dur := time.Duration((300 + rand.Intn(300))) * time.Millisecond
	rf.electionID += 1
	rf.nelectiontimeout += 1
	//if rf.nelectiontimeout > 2 {
		//log.Fatal("more than 2 election loops running")
	//}
	rf.setTimeout(rf.electionID, &rf.electionID, &rf.nelectiontimeout, dur, rf.startElectionCycle)
}

func (rf *Raft) startHearbeatTimeout(){
	dur := time.Duration((149 + rand.Intn(150))) * time.Millisecond
	rf.heartbeatID += 1
	rf.nheartbeattimeout += 1
	//if rf.nelectiontimeout > 2 {
		//log.Fatal("more than 2 heartbeat loops running")
	//}
	rf.setTimeout(rf.heartbeatID, &rf.heartbeatID, &rf.nheartbeattimeout, dur, rf.onHeartbeatTimeout)
}



type RequestVoteArgs struct {
	// id of node requesting the vote
	From int
	VoteChannel chan bool
	Term int
}

type RequestVoteReply struct {
	Vote bool
}

type AppendEntryReply struct {
}

func (rf *Raft) castVote(from int){
	rf.mu.Lock()
	defer rf.mu.Unlock()
	log.Println("Casting vode for node ", from, " at node ", rf.me)
	rf.hasvoted = true
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {

	if args.Term > rf.term {
		rf.mu.Lock()
		rf.term = args.Term
		rf.hasvoted = false
		rf.mu.Unlock()
		if rf.isLeader{
			rf.revokeLeadership()
		}
	} else if args.Term < rf.term {
		return
	}

	go rf.startElectionTimeout()

	if !rf.hasvoted{
		reply.Vote = true
	}

	rf.castVote(args.From)
}

func (rf *Raft) onHeartbeatTimeout(){
	for id, _ := range rf.peers {
		go func (id int) {
			if id == rf.me { return }
			ok := rf.peers[id].Call("Raft.AppendEntries", &AppendEntry{Term: rf.term}, &AppendEntryReply{})
			if !ok { // lost contact
				log.Println("lost contanct at node ", rf.me)
				//rf.revokeLeadership()
			}
		}(id)
	}
	go rf.startHearbeatTimeout()
}

func (rf *Raft) becomeLeader() {
	log.Println("Won election at node ", rf.me)
	rf.mu.Lock()
	rf.isLeader = true
	rf.electionID++
	rf.mu.Unlock()

	go rf.startHearbeatTimeout()
}

func (rf *Raft) collectVotes(votes chan bool){
	nvotes := 1 // voted for self
	for vote := range votes {
		if vote { nvotes++ } 
		if nvotes > (len(rf.peers)/2 + len(rf.peers) % 2) {
			rf.becomeLeader()
			break
		}
	}
	if !rf.isLeader{
		close(votes)
		rf.hasvoted = false
		go rf.startElectionTimeout()
	}
}

func (rf *Raft) startElectionCycle(){
		log.Println("requesting votes for Node ", rf.me)

		go rf.startElectionTimeout()
		rf.mu.Lock()
		rf.term++
		rf.mu.Unlock()
		rf.castVote(rf.me)


		voteChannel := make(chan bool, len(rf.peers)-1)

		for i, _ := range rf.peers{
			if i == rf.me{
				continue
			}
			go rf.sendRequestVote(i, &RequestVoteArgs{ From: rf.me, VoteChannel: voteChannel, Term: rf.term } , &RequestVoteReply{} )
		}

		go rf.collectVotes(voteChannel)

}




func (rf *Raft) AppendEntries(args *AppendEntry, reply *AppendEntryReply){
	log.Println("heartbeat received at ", rf.me)

	if args.Term >= rf.term {
		rf.term = args.Term
		if rf.isLeader{
			rf.revokeLeadership()
		}
	}

	go rf.startElectionTimeout()
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

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	args.VoteChannel <- reply.Vote
	return ok
}

//
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
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// begin agreement process on a new log entry
	// save command in a variable 
	// send a broadcast saying to start logging this
		// rpc call
			// logged it or couldnt 
	// get back from teh majority of nodes that they have logged the thing
	// store `commmand` on disk

	// notify the service of the commit

	rf.applychan <- ApplyMsg{
		CommandValid : true,
		Command      : command,
		CommandIndex : 0, // TODO:
	}

	// Your code here (2B).


	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	log.Println(rf.me, " has been killed")

	rf.revokeLeadership()
	//rf.electionID++

	//go func() {
		//for {
			//if rf.killed(){
				//time.Sleep(time.Duration(1) * time.Second)
			//}else {
				//break
			//}
		//}// not killed anymore
		//log.Println("reconnecting this node")
		//rf.startElectionTimeout()
	//}()

}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}



//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.applychan = applyCh

	rf.readPersist(persister.ReadRaftState())

	go rf.startElectionTimeout()

	return rf
}
