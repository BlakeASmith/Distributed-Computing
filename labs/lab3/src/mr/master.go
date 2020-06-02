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
import "time"

type Stat struct {
	TotalTime                   time.Duration
	StartTime                   time.Time
	EndTime                     time.Time
	MapSequenceTime             time.Duration
	ReduceSequenceTime          time.Duration
	StatsPerJob                 []TaskStat
	PartitioningTime            time.Duration
	PartitioningTimePerBucket   []time.Duration
	PartitioningSpace           []int
	IntermedateHDSpaceForMap    int
	IntermedateHDSpaceForReduce int
	NumberOfReduceJobs          int
	NumberOfMapJobs             int
	NumberOfUniqueKeys          int
	NumberOfRecords             int
	InputBytes                  int
}

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
	// performance stats
	stats Stat
}

func SizeOfFile(f *os.File) int {
	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}
	return int(stat.Size())
}

func (m *Master) StartReducePhase() {
	m.FlagMutex.Lock()
	m.Reducing = true
	m.FlagMutex.Unlock()

	m.stats.MapSequenceTime = time.Since(m.stats.StartTime)

	log.Println("Renaming temp files and creating reduce tasks")

	m.Jobs = make([]JobInfo, m.nReducers)
	for i, f := range m.ReduceBuckets {
		newname := fmt.Sprintf("mr_%d", i)
		os.Rename(f.Name(), newname)
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

func writebucket(m *Master, kvs []KeyValue, bucket *os.File, wg *sync.WaitGroup) {
	defer wg.Done()
	jsonstr := ""
	for _, kv := range kvs {
		str, err := json.Marshal(kv)
		if err != nil {
			log.Fatal(err)
		}
		jsonstr += string(str)
		jsonstr += "\n"
	}
	log.Println("writing to ", bucket.Name())
	bucket.WriteString(jsonstr)
}

func (m *Master) WriteToBuckets(rnum int, rtchan chan time.Duration,
	spchan chan int, wg *sync.WaitGroup) {

	defer wg.Done()

	start := time.Now()

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

	writewg := sync.WaitGroup{}
	writewg.Add(len(bins))

	for i, kvs := range bins {
		go writebucket(m, kvs, m.ReduceBuckets[i], &writewg)
	}

	log.Println("waiting for writebucket goroutines")
	writewg.Wait()
	log.Println("writing goroutines finished")

	memoryuse := 0
	for _, lst := range bins {
		for _, s := range lst {
			memoryuse += len(s.Key)
			memoryuse += len(s.Value)
		}
	}

	rtchan <- time.Since(start)
	spchan <- memoryuse

	return
}

// RPC for the worker to let us know it has finished
func (m *Master) ConfirmFinish(args *WorkerTask, reply *CompleteConfirmation) error {
	m.JobMutex.Lock()
	defer m.JobMutex.Unlock()
	if m.Jobs[args.Job].Complete {
		return errors.New("Job already finished")
	}

	m.stats.NumberOfRecords += args.Stat.Records
	m.stats.NumberOfUniqueKeys += args.Stat.UniqueKeys
	m.stats.StatsPerJob = append(m.stats.StatsPerJob, args.Stat)

	m.Jobs[args.Job].Complete = true
	log.Printf("Job: %d completed by %d", args.Job, args.WorkerNumber)

	done := true
	for _, job := range m.Jobs {
		if !job.Complete {
			done = false
		}
	}

	start := time.Now()
	if done && !m.Reducing {
		wg := sync.WaitGroup{}
		timechan := make(chan time.Duration, 100)
		spacechan := make(chan int, 100)

		wg.Add(len(m.Jobs))

		for i := 0; i < len(m.Jobs); i++ {
			go m.WriteToBuckets(i, timechan, spacechan, &wg)
		}

		log.Println("Waiting for goroutines")
		wg.Wait()
		close(timechan)
		close(spacechan)
		log.Println("goroutines finished")

		for sp := range spacechan {
			m.stats.PartitioningSpace = append(m.stats.PartitioningSpace, sp)
		}
		for rt := range timechan {
			m.stats.PartitioningTimePerBucket = append(m.stats.PartitioningTimePerBucket, rt)
		}

		log.Println("Starting Reduce")

		m.StartReducePhase()
	}
	m.stats.PartitioningTime = time.Since(start)

	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	m.stats.StartTime = time.Now()
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

// remove intermedate files and count bytes stored
func DoCleanup(m *Master) {
	m.ReduceMutex.Lock()
	defer m.ReduceMutex.Unlock()

	mapbytesbytask := make([]int, m.stats.NumberOfMapJobs)
	for i := 0; i < m.stats.NumberOfMapJobs; i++ {
		ifile, err := os.Open(fmt.Sprintf("map_%d", i))
		if err != nil {
			panic(err)
		}
		mapbytesbytask = append(mapbytesbytask, SizeOfFile(ifile))
		os.Remove(ifile.Name())
		ifile.Close()
	}

	tbytes := 0
	for _, bytes := range mapbytesbytask {
		tbytes += bytes
	}

	m.stats.IntermedateHDSpaceForMap = tbytes

	rbytes := 0
	for _, bucket := range m.ReduceBuckets {
		rbytes += SizeOfFile(bucket)
	}
	m.stats.IntermedateHDSpaceForReduce = rbytes
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
		m.stats.EndTime = time.Now()
		m.stats.TotalTime = time.Since(m.stats.StartTime)

		for _, stat := range m.stats.StatsPerJob {
			m.stats.NumberOfUniqueKeys += stat.UniqueKeys
			m.stats.NumberOfRecords += stat.Records
		}

		m.stats.ReduceSequenceTime = time.Since(m.stats.StartTime.Add(m.stats.MapSequenceTime))

		DoCleanup(m)

		// write stats to file
		_, err := os.Stat("stats.json")
		var f *os.File
		if os.IsNotExist(err) {
			f, err = os.Create("stats.json")
		} else {
			f, err = os.OpenFile("stats.json", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		}

		if err != nil {
			log.Fatal()
		}
		defer f.Close()

		_json, err := json.Marshal(m.stats)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(_json))
		if _, err := f.WriteString(string(_json) + "\n"); err != nil {
			log.Fatal(err)
		}
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

	// calculate input bytes
	ibytes := 0
	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		ibytes += SizeOfFile(f)
	}

	m.stats = Stat{}
	m.stats.InputBytes = ibytes

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
	m.stats.NumberOfReduceJobs = nReduce
	m.stats.NumberOfMapJobs = len(files)
	m.stats.PartitioningTimePerBucket = make([]time.Duration, nReduce)
	m.stats.PartitioningSpace = make([]int, nReduce)
	m.stats.StatsPerJob = make([]TaskStat, 0)
	m.FlagMutex = sync.Mutex{}
	m.JobMutex = sync.Mutex{}
	m.ReduceMutex = sync.Mutex{}

	m.server()
	return &m
}
