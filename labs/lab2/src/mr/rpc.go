package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

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

type WorkerTask struct {
	Filename     string
	Reduce       bool
	WorkerNumber int
	Job          int
	NReducers    int
	Finish       bool
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
