package raft

import "sync"
import "sync/atomic"
import "time"
import "log"
import "math/rand"
import "../labrpc"

// import "bytes"
// import "../labgob"

func (rf *Raft) log(content ...interface{}){
	log.Println(rf.me, ": ", content)
}

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	currentTerm int
	votedFor int
	votes int

	isleader bool
	currentLeader int

	electionID int
	heartbeatID int

	commitmu sync.Mutex
	apply chan ApplyMsg
	commitInd int
	lastApplied int
	_log []LogEntry

	nextIndex []int
	matchIndex []int
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.apply = applyCh


	rf.commitmu = sync.Mutex{}

	rf._log = make([]LogEntry, 0)

	rf.follower()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())


	return rf
}


type RequestVoteArgs struct {
	Term int
	ID int
	LastLogTerm int
	LogLen int
}

type RequestVoteReply struct {
	Vote bool
	PrevLogIndex int
	PrevLogTerm int
	Success bool
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	if rf.dead == 1 { return }
	//rf.log("recieved a vote request from ", args.ID)
	rf.mu.Lock()

	if args.Term > rf.currentTerm{
		rf.currentTerm = args.Term
	} else if args.Term < rf.currentTerm{
		//reply.Term = rf.currentTerm
		return
	}

	rf.mu.Unlock()
	rf.follower()
	rf.mu.Lock()

	if rf.votedFor == -1 || rf.votedFor == args.ID {
		rf.votedFor = args.ID
		reply.Vote = true
	}

	if len(rf._log) > 0 {
		llt := rf._log[len(rf._log)-1].Term
		if args.LastLogTerm < llt || (args.LastLogTerm == llt && len(rf._log) > args.LogLen)    {
			reply.Vote = false
		}
	}

	rf.mu.Unlock()
}

// perform an action after a specified timeout
// the action will not be perfimed if the given ID changes
func timeout(id int, currentID *int, duration time.Duration, doThing func() string) {
	time.Sleep(duration)
	if id == *currentID{
		log.Println("DOING THING ")
		doThing()
	}
}

func (rf *Raft) heartbeat() string {
	//if rf.killed() { return }
	rf.log("BEATING")

	for id, server := range rf.peers{
		if id != rf.me{
			server.Call(
				"Raft.AppendEntries",
				&AppendEntryArgs{
					Term: rf.currentTerm,
					LeaderID: rf.currentLeader,
					LeaderCommitIndex: rf.commitInd,
				},&AppendEntryReply{},
			)
		}
	}

	go rf.setHeartbeat()
	return "HEARTBEAT"
}

func (rf *Raft) setHeartbeat(){
	duration := time.Duration(50 + rand.Intn(150)) * time.Millisecond
	timeout(rf.heartbeatID, &rf.heartbeatID, duration, rf.heartbeat)
}

func (rf *Raft) leader(){
	if rf.dead == 1 { return }
	rf.mu.Lock()

	rf.log("in leader state")
	rf.isleader = true
	rf.currentLeader = rf.me
	rf.electionID += 1 // cancel the election timeout
	rf.votes = 0

	rf.nextIndex = make([]int, len(rf.peers))

	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = len(rf._log)
	}

	rf.matchIndex = make([]int, len(rf.peers))

	rf.mu.Unlock()
	go rf.setHeartbeat()
}

func (rf *Raft) candidate() string{
	//if rf.killed() { rf.log("KILL") ; return }
	rf.mu.Lock()
	rf.log("in candiate state")
	rf.currentTerm += 1
	rf.votes = 0
	rf.votes += 1 // vote for myself
	rf.votedFor = rf.me
	rf.mu.Unlock()

	rf.electionID += 1 // cancel any previous timers

	// start a new timer for the candidacy
	//duration := time.Duration(600 + rand.Intn(150)) * time.Millisecond
	//go timeout(rf.electionID, &rf.electionID, duration, rf.follower)

	for server, _ := range rf.peers {
		if server == rf.me {
			continue
		}

		llt := -1
		if len(rf._log) > 0{
			llt = rf._log[len(rf._log)-1].Term
		}

		args := RequestVoteArgs{
			Term: rf.currentTerm,
			ID: rf.me,
			LastLogTerm: llt,
			LogLen: len(rf._log),
		}

		reply := RequestVoteReply{}
		rf.sendRequestVote(server, &args, &reply)
		rf.mu.Lock()
		if reply.Vote {
			rf.votes += 1
		}
		rf.mu.Unlock()
	}

	rf.electionID += 1 // end candidacy timer

	if rf.votes >= (len(rf.peers)/2 + len(rf.peers) % 2){
		rf.log("ELECTED with ", rf.votes, " votes out of ", len(rf.peers))
		rf.leader()
	}


	return "CANDIDATE"
}

func (rf *Raft) receiveHeartbeat(args *AppendEntryArgs) {
	rf.log("recieved hearbeat")
	rf.currentLeader = args.LeaderID
	rf.mu.Lock()
	if args.Term >= rf.currentTerm{
		rf.currentTerm = args.Term
	}
	rf.mu.Unlock()
	rf.follower()
}

func (rf *Raft) electionTimer(){
	rf.electionID += 1 // cancel any previous timers
	duration := time.Duration(400 + rand.Intn(200)) * time.Millisecond
	go timeout(rf.electionID, &rf.electionID, duration, rf.candidate)
}

func (rf *Raft) follower() string{
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.dead == 1 { return "FOLLOWER" }
	rf.log("in follower state")
	rf.isleader = false
	rf.heartbeatID += 1 // cancels the heartbeat if the node was a leader
	rf.votedFor = -1 // ready to vote again
	rf.electionID += 1
	rf.electionTimer() // restart election timer
	return "FOLLOWER"
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.log("STATE: ", "term: ", rf.currentTerm, "isleader: ", rf.isleader)
	return rf.currentTerm, rf.isleader
}



func (rf *Raft) Kill() {
	rf.log("KILLED")
	rf.follower()
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.electionID += 1
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

type LogEntry struct {
	Command interface{}
	Term int
	Index int
}

type AppendEntryArgs struct {
	Term int
	LeaderID int
	PrevLogIndex int
	PrevLogTerm int
	Entries []LogEntry
	LeaderCommitIndex int
}

type AppendEntryReply struct {
	Term int
	Success bool
}

func writeToLog(rf *Raft, entries []LogEntry){
	rf.log("wrote to logs: ", len(entries))
	for _, entry := range entries {
		rf._log = append(rf._log, entry)
	}
}

func commitLogs(rf *Raft, start int, end int){
	for _, cmd := range rf._log[start:end]{
		applymsg := ApplyMsg{
			Command: cmd.Command,
			CommandValid: true,
			CommandIndex: rf.commitInd+1,
		}
		rf.log("committing commandIndex ", applymsg.CommandIndex)
		rf.apply <- applymsg
		rf.commitInd += 1
	}
}

func (rf *Raft) setUpAppendEntriesArgs() []AppendEntryArgs {
	args := make([]AppendEntryArgs, len(rf.peers))
	for id, _ := range rf.peers {
		if id == rf.me { args[id] = AppendEntryArgs{}; continue }
		args[id] = AppendEntryArgs{
			Term: rf.currentTerm,
			LeaderID: rf.me,
			PrevLogIndex: rf.nextIndex[id]-1,
			PrevLogTerm: func () int {
				if rf.nextIndex[id] == 0 {
					return -1
				} else {
					return rf._log[rf.nextIndex[id]-1].Term
				}
			}(),
			Entries: rf._log[rf.nextIndex[id]:],
			LeaderCommitIndex: rf.commitInd,
		}
	}
	return args
}


func (rf *Raft) sendEntries(args AppendEntryArgs, id int, collect chan bool) {
			success := false
			loglen := len(rf._log)
			for !success && rf.nextIndex[id] >= 0 {
				reply := AppendEntryReply{}
				rf.peers[id].Call("Raft.AppendEntries", &args, &reply)
				success = reply.Success
				if !success { rf.nextIndex[id]-- }
			}
			//rf.log("collected result from ", id)
			rf.nextIndex[id] = loglen
			collect <- success
}

func broadcastCommit(rf *Raft) {
	//for id, server := range rf.peers {
		//if id == rf.me { continue }
		//server.Call("Raft.AppendEntries", &AppendEntryArgs{
			//LeaderCommitIndex: rf.commitInd,
		//}, &AppendEntryReply{})
	//}
}

func collectAppendResults (rf *Raft, collect chan bool) {

	// TODO: Break on new election
	nSuccess := 1
	for range collect {
		nSuccess++
		if nSuccess >= ((len(rf.peers))/2 + ((len(rf.peers)) % 2)) {
			rf.log("MAJORITY")
			//go broadcastCommit(rf)
			commitLogs(rf, rf.commitInd, len(rf._log)) // COMMIT
			break
		}
	}

}

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	if !rf.isleader { return rf.commitInd+1, rf.currentTerm, rf.isleader }

	rf.log("starting log replication")

	writeToLog(rf, []LogEntry{ LogEntry{
		Command: command,
		Term: rf.currentTerm,
		Index: len(rf._log),
	} })

	collect := make(chan bool)
	args := rf.setUpAppendEntriesArgs()

	for id, _ := range rf.peers {
		if id == rf.me { continue }
		go rf.sendEntries(args[id], id, collect)
	}

	go collectAppendResults(rf, collect)


	return rf.commitInd+1, rf.currentTerm, rf.isleader
}

func deleteEntriesFromLog(rf *Raft, startInd int){
	rf.log("deleting logs")
	rf._log = rf._log[0:startInd+1]
}


func (rf *Raft) AppendEntries(leader *AppendEntryArgs, reply *AppendEntryReply) {
	//if rf.killed() { return }
	rf.log(leader.LeaderID, rf.votedFor)

	go rf.receiveHeartbeat(leader)

	if leader.Term < rf.currentTerm {
		rf.log("Inconsistent term ", leader.Term, rf.currentTerm)
		return
	}

	if leader.PrevLogIndex != -1 && len(leader.Entries) > 0 && rf._log[leader.PrevLogIndex].Term != leader.PrevLogTerm {
		deleteEntriesFromLog(rf, leader.PrevLogIndex)
		return
	} else {
		writeToLog(rf, leader.Entries)
		reply.Success = true
	}

	if leader.LeaderCommitIndex > rf.commitInd {
		if len(rf._log)-1 < leader.LeaderCommitIndex{
			commitLogs(rf, rf.commitInd, len(rf._log))
		} else {
			commitLogs(rf, rf.commitInd, leader.LeaderCommitIndex)
		}
	}
}


func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// drf.persister.SaveRaftState(data)
}


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

