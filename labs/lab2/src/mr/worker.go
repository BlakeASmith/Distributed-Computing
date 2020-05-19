package mr

import "fmt"
import "log"
import "net/rpc"
import "hash/fnv"
import "io/ioutil"
import "encoding/json"
import "os"

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

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
func Reduce(reducef func(string, []string) string) []KeyValue {
	log.Println("Not implemented")
	return nil
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	args := WorkRequest{}
	resp := WorkerTask{}

	log.Println("requesting a task from the Master")
	call("Master.RequestTask", &args, &resp)

	if resp.reduce {
		Reduce(reducef)
		// TODO: save reult someplace
	}

	filetemplate := "map_%s"
	tempfilenametemplate := fmt.Sprintf("temp_%s", filetemplate)

	log.Println("reading in the task content")
	filename := resp.Filename
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("performing the map operation on %s\n", filename)
	// perform the map and put values in itermediate file
	results := mapf(filename, string(content))
	tempfilename := fmt.Sprintf(tempfilenametemplate, filename)
	f, err := ioutil.TempFile(".", tempfilename)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("writing data into %s\n", tempfilename)
	encoder := json.NewEncoder(f)
	for _, kv := range results {
		encoder.Encode(&kv)
	}
	// rename the temp file once we know the job is done
	newname := fmt.Sprintf(filetemplate, filename)
	log.Printf("renaming %s to %s\n", tempfilename, newname)
	err = os.Rename(f.Name(), newname)
	if err != nil {
		log.Fatal(err)
	}
}

//
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
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
