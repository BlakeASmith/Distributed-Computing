Lab 2: A Simplified Distributed Map Reduce
==============================================

In this lab we implemented a "distributed" map reduce 
(relies on a shared file system but otherwise runs independently). 
There is one "master" process handling the assignment of map and reduce jobs. 

During the mapping phase the Worker
process reads in a text file and produces a 
set of key value pairs using a map function provided by 
a plug-in. The results from the map function are stored in an intermediary
file from which the master can make files for the reduce tasks. 

Once all the mapping is complete the Master goes through 
the files produced from the mapping phase 
and produces a file for each reduce task that will be run.

Finally for each reduce job a Worker process groups 
the values by their keys, producing 
a :obj:`map[string][]string`. Then the reduce function 
provided by the plug-in will be called  on each 
slice of strings and it's associated key. 
The resulting value will be stored with it's key in
an output file.

Combining the results from all the reduce tasks is handled 
by the provided template code.

.. note::
        This was my first time working with go so I may have done some 
        silly things in my implementation. I'm sure i'll come to see 
        the issues in the coming weeks.

Basic Approach
-----------------------------------------

1. Worker processes invoke a RPC :func:`RequestTask` to receive 
   a job assignment from the master
2. Once a Worker completes a job it invokes another RPC :func:`ConfirmComplete`
   to inform the Master that the job has been finished
3. Workers continue to ask for assignments with :func:`RequestTask` 
   until a :obj:`Finished` flag is set in the response indicating there are
   no more jobs
4. When all map jobs are complete a set of reduce jobs will be created. 
   A mutex lock will be held while preparing the files and creating the reduce 
   jobs so any Workers requesting a task at that time will hang until that 
   mutex is unlocked (this eliminates the need for explicit waiting)
5. Once reduce jobs are finished we will return true in the :func:`Done`
   which is called by a goroutine in the Master process every second (provided in the template code)

Implementation
-------------------------------------------

Type Definitions
__________________________________________

I created types for the following purposes:

:WorkerTask: a struct containing information about a job for the worker 
             to use

:CompleteConfirmation: Information returned from :func:`ConfirmComplete`.
                       It's used to tell the Worker thread to wait when 
                       necessary 

:SortedKV: A KeyValue pair that has a reduce task number (Bin number)
           associated to it. Used to separate the results from map jobs
           into files for the reduce tasks

:JobInfo: Information about a job for the Master process

:Master: The struct for the Master process. Contains global state
         and Mutexes.

.. code-block:: go
        :caption: Definition from rpc.go

        // KeyValue type provided in the template
        type KeyValue struct {
                Key   string
                Value string
        }

        // Infromation for a Worker process about a Job
        type WorkerTask struct {
                Filename     string
                Reduce       bool
                WorkerNumber int
                Job          int
                NReducers    int
                Finish       bool
        }

        // Confirmation provided to a worker after they 
        // report a job as finished
        type CompleteConfirmation struct {
                Wait bool
        }

        // Used to designate which reduce task should handle 
        // this key value pair (Bin is the reduce task number which 
        // should process it, it's computed using a hash function)
        type SortedKV struct {
                KV  KeyValue
                Bin int
        }

.. code-block:: go
        :caption: Definitions from master.go

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

The Worker Process 
__________________________________________

The code for the Worker consists mainly of 3 important functions

.. function:: Reduce(reducef func(string, []string) string)

        handles processing a reduce job

.. function:: Map(mapf func(string, string) []Keyvalue)

        process a map job

.. function:: Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) 

        The main entry point for a 
        Worker process to invoked by the template code.

        :param func mapf: The map function from the plugin

        :param func reducef: the reduce function from the plugin

The Worker Function
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The :func:`Worker` function calls the :func:`RequestTask`
RPC from the Master, invokes the appropriate handler
(:func:`Map` for :func:`Reduce`) and then informs the 
Master of completion using the :func:`ConfirmComplete`
RPC

.. code-block:: go

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

                        if resp.Reduce {
                                Reduce(resp, reducef)
                        } else {
                                Map(resp, mapf)
                        }

                        log.Println("Confirming completion to the master")
                        confirm := CompleteConfirmation{}
                        call("Master.ConfirmFinish", &resp, &confirm)

                        if confirm.Wait {
                                time.Sleep(time.Second)
                        }
                }
        }

The Map Function
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The :func:`Map` function is responsible for running a map job
and producing an intermediate file

.. code-block:: go

        func Map(task WorkerTask, mapf func(string, string) []KeyValue) {
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
                err = os.Rename(f.Name(), newname)
                if err != nil {
                        log.Fatal(err)
                }
                return
        }

The Reduce Function
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The :func:`Reduce` function handles the reduce jobs

.. code-block:: go

        func Reduce(task WorkerTask, reducef func(string, []string) string) {
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
                if err != nil {
                        log.Fatal(err)
                }
                //log.Println("Writing redults to ", outputfile.Name())
                for _, kv := range reduceresults {
                        outputfile.WriteString(
                                fmt.Sprintf("%v %v\n", kv.Key, kv.Value))
                }
                return
        }

The Master Process
-------------------------------------------

The important functions for the Master are: 

.. function:: Master.RequestTask(args `*workreqest`, reply `*WorkerTask`)

        Handles assignment of jobs 

.. function:: Master.ConfirmFinish(mapf func()

        Handles completed jobs

.. function:: Master.StartReducePhase() 

        Creates reduce jobs from the intermediate map files

:func:`RequestTask` finds the next available task and supplies the 
task information to the Worker

.. code-block:: go
        :caption: RequestTask

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

:func:`ConfirmFinish` marks the job as done and initiates the
reduce phase once all the map jobs are finished

.. code-block:: go
        :caption: ConfirmFinish
        
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

:func:`StartReducePhase` creates the reduce jobs so their ready to be
assigned by :func:`RequestTask`

.. code-block:: go
        :caption: StartReducePhase

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

The rest of the system is implemented in the provided template code


Limitations
-----------------------------------

Dev Log
-----------------------------

1. Initialized Master with 2 Maps. One to keep track of which files are currently being mapped and 
   another to keep track of which files have been successfully mapped 

.. note::

        Likley will need to use a Mutex for the Maps as Go maps are not threadsafe

2. Created a rpc :func:`RequestTask()` which workers can call to receive a filename & a boolean indicating
   whether to act as a mapper or a reducer

3. Implemented the Mapping functionallity in the worker
   - Logged Json serialization of each KeyValue returned from mapf in a TempFile
   - Rename temp file once all KeyValues have been properly written

:End of Lab: Had some problems with the go import/build system which took up about 1/2 hour of the lab.
             Otherwise I'm happy with the progress

:Next Steps: 1. When the code in invoked from the test script the temp files cannot be found properly. 
                using absolute paths will probably fix this

                * FIXED: needed to remove "../" from the filename received from master

             3. The shared map will probably cause race conditions. Atleast wrap the access in 
                a function so it'll be easy to fix this later

                * FIXED: Added a Mutex to the Master struct for this purpose

             4. Create an RPC for workers to call when whey have finished.
                This RPC can then handle determining if we are Done or not, 
                once mapping is done :func:`RequestTask()` can start handing outreduce tasks

                * FIXED: Added an RPC for the worker to call when they finish
                * The Files map is updated so the main RPC will know when
                  to start handing out Reduce Tasks

             6. Create an id/numbering system for the workers, You will need this to make sure the same
                file is not mapped twice (if a worker finishes after the timeout we will need to tell
                the difference between it and the newly dispatched worker)

                * DONE

             5. Create and start a GoRoutine to keep track of how long
                workers have been running and free the filenames to be given
                out again by :func:`RequestTask()` when the workers take to long (10s)

                * DONE: Was able to use the Done() routine which is already called
                  every second

             6. Assign Reduce tasks after the Map tasks have completed

                * DONE: The list of input files is split evenly umong the
                  available workers and provided on the next request to 
                  :func:`RequestTask()`




:TODO: Assign Reduce Tasks

       * DONE:

:TODO: Put loop in workers to continually request tasks

       * DONE: 


:Bug: Test hangs after first 3 maps

      * Nevermind, not a bug. The test only runs
        3 workers

:Bug: Tests pass with 2 workers but not with 3. Reduce buckets are being written to after
      the reduce tasks have been started

      * FIXED: Tasks were being marked complete when the started processing 
                not after finishing

:Issue: Producing the reduce files from each of the worker threads output files is 
        really slow

:Note: I had to move the bin sorting to the end of the map phase before the reduce phase. 
       It seems the order that these are done is important but I don't understand why. A Mutex
       is used so there should not be any concurrency issues

:Bug: Hangs on crash.go. It gets to the reduce tasks fine but eventually 
      tasks stop being handed out without the code finishing. 
      The Done function checks that all tasks are finished and then returns true, 
      so it should not be possible for the code to hang unless there are still some tasks
      not marked as complete. Perhaps tasks are not being re-assigned properly
      when a reducer fails mid shift

        * FIXED: I was looping over the jobs using range. So when the assigned & complete variables were
          changed in the loop it was a copy that was changed and not the real value. By getting the 
          correct address of each job and using that the issue was fixed


:DONE:
        All tests passing as of May 24th 2020. Also tested with up to 6 worker processes 



