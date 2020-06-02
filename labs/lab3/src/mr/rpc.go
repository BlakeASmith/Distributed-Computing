package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"
import "time"

// struct to contain all stat info
type Stats struct {
	// total execution time
	TotalTime time.Duration
	// time to produce intermedates
	MapSeqenceTime time.Duration
	// (in-memory) space required to store intermediates (in bytes)
	IntermediateSpace int
	// time to sort the intermediates
	SortTime time.Duration
	// runtime of reduce jobs, including storage of the results
	ReduceSequenceTime time.Duration
	// number of bytes from all the keys (including duplicates)
	NumKeyBytes int
	// number of bytes from all the values
	NumValBytes int
	// number of KV pairs processed
	NumRecords int
	// number of keys after grouping (number of calls to reduce
	NumKeys int
}

var stats = &Stats{}

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type KeyValue struct {
	Key   string
	Value string
}

// Add your RPC definitions here.

// Don't know what im putting in here yet
type WorkRequest struct {
	Arg1 string
}

type TaskStat struct {
	JobNumber  int
	TaskID     int
	Reduce     bool
	Time       time.Duration
	Space      int
	UniqueKeys int
	Records    int
}

type WorkerTask struct {
	Filename     string
	Reduce       bool
	WorkerNumber int
	Job          int
	NReducers    int
	Finish       bool
	Stat         TaskStat
}

type SortedKV struct {
	KV  KeyValue
	Bin int
}
type CompleteConfirmation struct {
	Wait bool
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
