package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "errors"

type Master struct {
	Files     map[string]bool
	Finished  map[string]bool
	nReducers int
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) RequestTask(args *WorkRequest, reply *WorkerTask) error {
	for filename, active := range m.Files {
		if !active {
			reply.Filename = filename
			// NOTE: this access is not threadsafe
			m.Files[filename] = true
			return nil
		}
	}
	return errors.New("No jobs available")
}

func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	return ret
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}
	// a map to keep track of which files have been assigned
	// to workers NOTE: May need to use a Mutex for the map
	m.Files = make(map[string]bool)
	for _, file := range files {
		m.Files[file] = false
	}
	m.nReducers = nReduce

	m.server()
	return &m
}
