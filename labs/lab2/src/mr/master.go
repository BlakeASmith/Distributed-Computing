package mr

import "log"
import "net"
import "os"
import "fmt"
import "io/ioutil"
import "encoding/json"
import "net/rpc"
import "net/http"
import "errors"
import "sync"

type JobInfo struct {
	// The Worker number that the task has been asssigned to
	Assigned int
	// Flag to indicate that the task is finished
	Complete bool
	// name of the file being mapped/reduced
	Filename string
	// flag indicating a Reduce job
	Reducing bool
	// Counter used to determine that the task has ran too long
	RunTime int
}

type Master struct {
	// The list of jobs
	Jobs     []JobInfo
	JobMutex sync.Mutex
	// The reduce files
	ReduceBuckets []*os.File
	// Mutex for the master
	ReduceMutex sync.Mutex
	// Number of reduces
	nReducers int
	// Flag to indicate that we are in the reduce phase
	Reducing bool
	// counter used to assign ids to tasks
	taskcount int
	FlagMutex sync.Mutex
}

func (m *Master) StartReducePhase() {
	m.FlagMutex.Lock()
	m.Reducing = true
	m.FlagMutex.Unlock()

	log.Println("Renaming temp files and creating reduce tasks")

	m.Jobs = make([]JobInfo, m.nReducers)
	for i, f := range m.ReduceBuckets {
		newname := fmt.Sprintf("mr_%d", i)
		os.Rename(f.Name(), newname)
		f.Close()
		m.Jobs[i] = JobInfo{
			Filename: newname,
			Reducing: true,
		}
		log.Println("added job ", m.Jobs[i])
	}

	return
}

func (m *Master) RequestTask(args *WorkRequest, reply *WorkerTask) error {
	m.JobMutex.Lock()
	defer m.JobMutex.Unlock()
	m.FlagMutex.Lock()
	defer m.FlagMutex.Unlock()

	for i := 0; i < len(m.Jobs); i++ {
		job := &m.Jobs[i]
		if job.Assigned == 0 {
			reply.Filename = job.Filename
			m.taskcount++
			log.Printf("assigned job %d, %s to worker %d\n",
				m.taskcount, reply.Filename, m.taskcount)
			reply.WorkerNumber = m.taskcount
			job.Assigned = m.taskcount
			reply.NReducers = m.nReducers
			reply.Job = i
			if m.Reducing {
				reply.Reduce = true
			}
			return nil
		}
	}

	if m.Reducing {
		reply.Finish = true
	}
	return nil
}

func (m *Master) WriteToBuckets(rnum int) {

	//log.Printf("Opening intermediate file for job %d\n", args.Job)
	file, err := os.Open(fmt.Sprintf("map_%d", rnum))
	if err != nil {
		log.Fatal(err)
	}
	//log.Printf("Decoding JSON data for job %d\n", args.Job)
	decoder := json.NewDecoder(file)
	bins := make([][]KeyValue, m.nReducers)
	for {
		var skv SortedKV
		if err := decoder.Decode(&skv); err != nil {
			break
		}
		bins[skv.Bin] = append(bins[skv.Bin], skv.KV)
	}

	log.Println("Writing pairs to appropriate Reducer files", rnum)

	m.ReduceMutex.Lock()
	defer m.ReduceMutex.Unlock()
	for i, kvs := range bins {
		jsonstr := ""
		for _, kv := range kvs {
			str, err := json.Marshal(kv)
			if err != nil {
				log.Fatal(err)
			}
			jsonstr += string(str)
			jsonstr += "\n"
		}
		m.ReduceBuckets[i].WriteString(jsonstr)
	}
	return
}

// RPC for the worker to let us know it has finished
func (m *Master) ConfirmFinish(args *WorkerTask, reply *CompleteConfirmation) error {
	m.JobMutex.Lock()
	defer m.JobMutex.Unlock()
	if m.Jobs[args.Job].Complete {
		return errors.New("Job already finished")
	}

	m.Jobs[args.Job].Complete = true
	log.Printf("Job: %d completed by %d", args.Job, args.WorkerNumber)

	done := true
	for _, job := range m.Jobs {
		if !job.Complete {
			done = false
		}
	}

	if done && !m.Reducing {
		for i := 0; i < len(m.Jobs); i++ {
			m.WriteToBuckets(i)
		}
		m.StartReducePhase()
	}

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
	log.Println("Checking if Done")
	jobsdone := true
	// Check for long running jobs
	m.JobMutex.Lock()
	defer m.JobMutex.Unlock()
	for i := 0; i < len(m.Jobs); i++ {
		job := &m.Jobs[i]
		if job.Assigned != 0 && !job.Complete {
			job.RunTime++
		}
		if job.RunTime >= 10 {
			log.Println("releasing job from ", job.Assigned)
			job.RunTime = 0
			job.Assigned = 0
			job.Complete = false
		}

		// check for unfinished jobs
		if !job.Complete {
			jobsdone = false
		}
	}

	m.FlagMutex.Lock()
	defer m.FlagMutex.Unlock()
	if jobsdone && m.Reducing {
		return true
	}

	return false
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//

func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// create a list of jobs from the input files
	m.Jobs = make([]JobInfo, len(files))
	for i, file := range files {
		m.Jobs[i] = JobInfo{Filename: file}
	}

	// open temorary files for each reduce bucket
	m.ReduceBuckets = make([]*os.File, nReduce)
	os.Mkdir("reducebuckets", 0755)
	tempnametemplate := "temp_%d"
	for i := 0; i < nReduce; i++ {
		tempname := fmt.Sprintf(tempnametemplate, i)
		f, err := ioutil.TempFile("reducebuckets", tempname)
		if err != nil {
			log.Fatal(err)
		}
		m.ReduceBuckets[i] = f
	}

	m.nReducers = nReduce
	m.FlagMutex = sync.Mutex{}
	m.JobMutex = sync.Mutex{}
	m.ReduceMutex = sync.Mutex{}

	m.server()
	return &m
}
