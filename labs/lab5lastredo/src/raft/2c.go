package raft

 import "bytes"
 import "../labgob"

type PersistentState struct {
	Log []LogEntry
	CommitInd int
	Term int
	VotedFor int
}

func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:

	 persisted := PersistentState{
		 Log: rf.Log._log,
		 CommitInd: rf.Log.commitInd.get(),
		 Term: rf.Term.get(),
		 VotedFor: rf.VotedFor.get(),
	 }

	 w := new(bytes.Buffer)
	 e := labgob.NewEncoder(w)
	 e.Encode(persisted)
	 data := w.Bytes()
	 rf.persister.SaveRaftState(data)
}


func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	 r := bytes.NewBuffer(data)
	 d := labgob.NewDecoder(r)
	 var persisted  PersistentState
	 if err := d.Decode(&persisted); nil != err{
		panic (err)
	 } else {
		rf.Log.commitInd.set(persisted.CommitInd)
		rf.Log._log = persisted.Log
		rf.VotedFor.set(persisted.VotedFor)
		rf.Term.set(persisted.Term)
	 }
}

