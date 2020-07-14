package raft





func (rf *Raft) AppendEntries(leader *AppendEntryArgs, reply *AppendEntryReply) {
	if rf.killed() { return }
	if(rf.me == leader.LeaderID) { return }
	if leader.Term < rf.Term.get() { return }

	reply.ID = rf.me
	rf.receiveHeartbeat(leader)
	rf.checkTerm(leader.Term)
	rf.Log.commitTo(MinOf(leader.LeaderCommitIndex-1, rf.Log.Size()), rf.me)

	if len(leader.Entries) > 0 {
		if rf.Log.PrevLogTerm() != leader.PrevLogTerm { return }
		if ((rf.Log.Size() + len(leader.Entries))) < leader.LogLen {
			//rf.log(leader.Entries, "AFAWFWAF<Plug>_")
			return
		}
		rf.log("writing ", leader.Entries)
		rf.Log.extend(leader.Entries)
		rf.log(rf.Log._log)
		reply.Success = true
	}


}

func commitInfo(rf *Raft) (int, int, bool){
	return rf.Log.getCommitInd(), rf.Term.get(), rf.LeaderID.get() == rf.me
}

func (rf *Raft) makeArgs(id int) AppendEntryArgs{
		prevLogIndex := rf.nextIndex[id].get()-1
		prevLogTerm := -1
		leaderCommitIndex := rf.Log.getCommitInd()
		if rf.nextIndex[id].get()-1 >= 0 && prevLogIndex < rf.Log.Size() {
			prevLogTerm = rf.Log._log[prevLogIndex].Term
		}

		//assert(leaderCommitIndex >= rf.nextIndex[id].get(), "")
		if rf.nextIndex[id].get() >= rf.Log.Size(){
			rf.nextIndex[id].set(rf.Log.Size()-1)
		}

		return AppendEntryArgs{
			Term: rf.Term.get(),
			LeaderID: rf.LeaderID.get(),
			PrevLogTerm: prevLogTerm,
			PrevLogIndex: prevLogIndex,
			Entries: rf.Log._log[rf.nextIndex[id].get():],
			LeaderCommitIndex: leaderCommitIndex,
			LogLen: rf.Log.Size(),
		}
}


func (rf *Raft) Start(command interface{}) (int, int, bool) {
	if !(rf.LeaderID.get() == rf.me) { return commitInfo(rf) }
	//rf.mu.Lock()
	//defer rf.mu.Unlock()
	loglen := rf.Log.Size()
	entry := LogEntry{
		Command: command,
		Term: rf.Term.get(),
		Index: loglen,
	}

	rf.log("writing at leader ", entry)
	rf.Log.add(entry)
	assert(rf.Log.Size() == loglen+1, "More than one item appended to log")
	rf.log(rf.Log._log)

	args := make([]AppendEntryArgs, len(rf.peers))
	loglen = rf.Log.Size()
	for id, _ := range args {
		args[id] = rf.makeArgs(id)
		args[id].LogLen = loglen
	}

	results := rf.sendAppendEntries(func (id int) AppendEntryArgs {
		return args[id]
	})

	go rf.collectSuccess(results)

	return commitInfo(rf)
}

func (rf *Raft) collectSuccess(results chan AppendEntryReply){
	loglen := rf.Log.Size()
	nSuccess := 0
	for reply := range results{
		if(reply.Success) {
			nSuccess++
		}
		if majority(nSuccess, len(rf.peers)-1){
			break
		}
	}
	rf.Log.commitTo(loglen, rf.me)
}

func (rf *Raft) initLeaderForLogging(){
	rf.nextIndex = make([]LockedInt, len(rf.peers))
	loglen := rf.Log.Size()
	rf.log("setting nextIndex ", rf.Log.commitInd.get())
	for i, _ := range rf.nextIndex{
		rf.nextIndex[i].set(loglen)
	}
}

