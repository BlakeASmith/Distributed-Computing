package raft

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {

	rf.checkTerm(args.Term)
	reply.Term = rf.Term.get()

	if args.Term < rf.Term.get(){
		return
	}
	if args.LastLogTerm < rf.Log.PrevLogTerm() ||
	args.LastLogTerm == rf.Log.PrevLogTerm() &&
	args.LogLen < rf.Log.Size() || args.CommitInd < rf.Log.commitInd.get(){
		rf.log("I have a better log, voting false to ", args.ID)
		return
	}

	if rf.VotedFor.get() == -1 || rf.VotedFor.get() == args.LeaderID || rf.VotedFor.get() == rf.me{
		reply.Vote = true
		rf.VotedFor.set(args.ID)
	}


}

func (rf *Raft) checkTerm(term int){
	if term > rf.Term.get(){
		rf.LeaderID.set(-1)
		rf.VotedFor.set(-1)
		rf.Term.set(term)
	}
}


func (rf *Raft) won(){
	rf.log("VICTORY")
	rf.LeaderID.set(rf.me)
	rf.doHeartbeat()
	rf.heartbeatTimer.reset()
	rf.initLeaderForLogging()
}

func (rf *Raft) startElection() {
	rf.log("starting election")
	rf.Term.inc(); rf.VotedFor.set(rf.me)

	votes := make(chan bool, len(rf.peers)-1)
	loglen := rf.Log.Size()
	prevLogTerm := rf.Log.PrevLogTerm()
	forEachPeerAsync(rf, func(ID int){
		reply := RequestVoteReply{}
		args := RequestVoteArgs{
			ID: rf.me,
			LeaderID: rf.LeaderID.get(),
			Term: rf.Term.get(),
			LogLen: loglen,
			LastLogTerm: prevLogTerm,
			CommitInd: rf.Log.commitInd.get(),
		}
		for {
			ok := rf.sendRequestVote(ID, &args, &reply)
			if ok { break }
		}
		rf.log("recieving vote ", reply.Vote, " from ", ID)
		votes <- reply.Vote
		rf.checkTerm(reply.Term)
	})

	nvote := 1
	nnovote := 0

	for vote := range votes {
		if vote {
			nvote++
		} else {
			nnovote++
		}

		if majority(nvote, len(rf.peers)){ // +
			rf.won()
			break

		} else if majority(nnovote, len(rf.peers)){
			// cannot win a majority
			break
		}
	}
	//close(votes)
}

func (rf *Raft) receiveHeartbeat(leader *AppendEntryArgs){
	rf.electionTimer.reset()

	if rf.LeaderID.get() == rf.me{
		rf.log("lost leadership")
		rf.heartbeatTimer.cancel()
		rf.LeaderID.set(leader.LeaderID)
		rf.VotedFor.set(-1)
	}
}
