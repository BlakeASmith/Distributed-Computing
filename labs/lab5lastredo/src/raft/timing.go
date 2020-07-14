package raft

import "time"

type AsyncTimeout struct {
	counter LockedInt
	duration func() time.Duration
	action func()
	repeating bool
}

func (t *AsyncTimeout) set() {
	go func (counterValue int) {
		repeat := true
		for repeat {
			time.Sleep(t.duration())
			if counterValue == t.counter.get(){
				t.action()
			}else {
				break
			}
			repeat = t.repeating
		}
	}(t.counter.get())
}

func (t *AsyncTimeout) cancel() {
	t.counter.inc()
}

func (t *AsyncTimeout) reset() {
	t.cancel()
	t.set()
}

func (rf *Raft) doHeartbeat() {
	rf.sendAppendEntries(func (id int) AppendEntryArgs {
		return AppendEntryArgs{
			LeaderID: rf.LeaderID.get(),
			Term: rf.Term.get(),
			LeaderCommitIndex: rf.Log.getCommitInd(),
		}
	})
}
