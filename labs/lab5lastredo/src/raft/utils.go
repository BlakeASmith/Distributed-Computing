package raft

import "log"
import "sync"
import "time"
import "errors"

func assert(conditon bool, message string){
	if !conditon{
		panic(errors.New(message))
	}
}

func (rf *Raft) log(content ...interface{}){
	log.Println(rf.me, ": ", content)
}

func forEachPeerAsync(rf *Raft, action func(ID int)){
	for ID, _ := range rf.peers {
		if ID != rf.me{
			go action(ID)
		}
	}
}

func runLocking(lock *sync.Mutex, action func()){
	lock.Lock()
	defer lock.Unlock()
	action()
}

func isClosed(ch <-chan bool) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

// perform an action after a specified timeout
// the action will not be perfimed if the given ID changes
func timeout(cancelCh chan bool, duration time.Duration, action func()) {
	time.Sleep(duration)
	select {
		case <-cancelCh: log.Println("TIMER CANCELLED")
		default: action()
	}
}

func (rf *Raft) sendAppendEntries(make_args func(id int) AppendEntryArgs) chan AppendEntryReply{
	replies := make(chan AppendEntryReply, len(rf.peers)-1)
	forEachPeerAsync(rf, func (ID int){
		//assert(rf.me == rf.LeaderID.get(), "non leader sending heartbeats")
		if rf.me != rf.LeaderID.get() {return}
		args := make_args(ID)
		reply := AppendEntryReply{}
		for {
			ok := rf.peers[ID].Call("Raft.AppendEntries", &args, &reply)
			if reply.Success || len(args.Entries) == 0 { break }
			if ok && rf.nextIndex[ID].get() > 0 {
				//rf.log("cant reach ", ID, "decrease nextIndex and try again ", reply.ID)
				rf.nextIndex[ID].dec()
				args = rf.makeArgs(ID)
				//rf.log(args)
			}
		}
		rf.checkTerm(reply.Term)
		rf.nextIndex[ID].plus(len(args.Entries))
		//rf.nextIndex[ID].set(rf.Log.Size())
		replies <- reply
	})
	return replies
}

func majority(votes int, size int) bool {
	return (votes >= size / 2  + (size % 2))
}

func MinOf(vars ...int) int {
    min := vars[0]

    for _, i := range vars {
        if min > i {
            min = i
        }
    }

    return min
}
