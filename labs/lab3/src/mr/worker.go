package mr

import "fmt"
import "log"
import "net/rpc"
import "hash/fnv"
import "io/ioutil"
import "encoding/json"
import "os"

//import "strings"
import "sort"

import "time"

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//function to do the reduce operation for Worker
func Reduce(task WorkerTask, reducef func(string, []string) string, stat *TaskStat) {
	bin, err := os.Open(task.Filename)
	if err != nil {
		log.Fatal(err)
	}
	//log.Println("reading JSON from ", task.Filename)
	decoder := json.NewDecoder(bin)
	kvs := make([]KeyValue, 0)
	for {
		var kv KeyValue
		if err := decoder.Decode(&kv); err != nil {
			break
		}
		kvs = append(kvs, kv)
	}

	stat.Records = len(kvs)

	//log.Println("grouping data by key for", task.Filename)
	bykey := make(map[string][]string, 0)
	for _, kv := range kvs {
		bykey[kv.Key] = append(bykey[kv.Key], kv.Value)
	}

	//log.Println("running reduce on each key for ", task.Filename)
	reduceresults := make([]KeyValue, 0)
	for key, vals := range bykey {
		reduced := reducef(key, vals)
		kv := KeyValue{Key: key, Value: reduced}
		reduceresults = append(reduceresults, kv)
	}

	sort.Sort(ByKey(reduceresults))

	outputfile, err := os.Create(fmt.Sprintf("mr-out-%d", task.Job))
	defer outputfile.Close()
	if err != nil {
		log.Fatal(err)
	}
	//log.Println("Writing redults to ", outputfile.Name())
	for _, kv := range reduceresults {
		outputfile.WriteString(
			fmt.Sprintf("%v %v\n", kv.Key, kv.Value))
	}

	filestat, err := outputfile.Stat()
	if err != nil {
		panic(err)
	}
	stat.Space = int(filestat.Size())
	stat.UniqueKeys = len(reduceresults)
	return
}

func Map(task WorkerTask, mapf func(string, string) []KeyValue, stat *TaskStat) {

	filename := task.Filename
	//log.Printf("reading in the task content from %s\n", filename)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	//log.Printf("creating a temporary file for %s", filename)
	tempfilename := fmt.Sprintf("temp_map_%s", task.Job)
	if err != nil {
		log.Fatal(err)
	}
	f, err := ioutil.TempFile(".", tempfilename)
	if err != nil {
		log.Fatal(err)
	}

	//log.Printf("performing the map operation on %s\n", filename)
	// perform the map and put values in itermediate file

	results := mapf(filename, string(content))
	//log.Printf("writing data into %s\n", tempfilename)

	encoder := json.NewEncoder(f)
	for _, kv := range results {
		skv := SortedKV{KV: kv, Bin: ihash(kv.Key) % task.NReducers}
		encoder.Encode(&skv)
	}
	// rename the temp file once we know the job is done
	newname := fmt.Sprintf("map_%d", task.Job)
	log.Printf("renaming %s to %s\n", tempfilename, newname)
	filestat, err := f.Stat()
	if err != nil {
		panic(err)
	}
	stat.Space = int(filestat.Size())
	stat.Records = len(results)
	err = os.Rename(f.Name(), newname)
	if err != nil {
		log.Fatal(err)
	}
	return
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	for {
		args := WorkRequest{}
		resp := WorkerTask{}
		log.Println("requesting a task from the Master")
		call("Master.RequestTask", &args, &resp)

		if resp.Finish {
			log.Println("quitting worker")
			return
		}

		if resp.Filename == "" {
			log.Println("sleeping due to empty filename")
			time.Sleep(time.Second)
			continue
		}

		start := time.Now()

		if resp.Reduce {
			resp.Stat.Reduce = true
			Reduce(resp, reducef, &resp.Stat)
		} else {
			Map(resp, mapf, &resp.Stat)
		}

		resp.Stat.Time = time.Since(start)
		resp.Stat.JobNumber = resp.Job
		resp.Stat.TaskID = resp.WorkerNumber

		log.Println("Confirming completion to the master")
		confirm := CompleteConfirmation{}
		call("Master.ConfirmFinish", &resp, &confirm)

		if confirm.Wait {
			time.Sleep(time.Second)
		}
	}

}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	//c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
