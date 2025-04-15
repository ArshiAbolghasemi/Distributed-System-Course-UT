package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
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

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	workerID := os.Getpid()

	for {
		args := TaskArgs{WorkerID: workerID}
		var reply TaskReply

		ok := call("Coordinator.GetTask", &args, &reply)
		if !ok {
			log.Println("RPC call GetTask failed, exiting")
			return
		}

		switch reply.TaskType {
		case TaskTypeMap:
			log.Printf("Worker %d received Map task: %d, file: %s\n", workerID, reply.TaskNumber, reply.FileName)
			doMapTask(reply.TaskNumber, reply.FileName, reply.NReduce, mapf)
			reportTask(TaskTypeMap, reply.TaskNumber, workerID)
		case TaskTypeReduce:
			log.Printf("Worker %d received Reduce task: %d\n", workerID, reply.TaskNumber)
			doReduceTask(reply.TaskNumber, reply.NMap, reducef)
			reportTask(TaskTypeReduce, reply.TaskNumber, workerID)
		case TaskTypeWait:
			time.Sleep(time.Second)
		case TaskTypeExit:
			log.Printf("Worker %d received Exit signal. Exiting.\n", workerID)
			return
		}
	}
}

func doMapTask(taskNumber int, filename string, nReduce int, mapf func(string, string) []KeyValue) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("doMapTask: unable to read file %v: %v", filename, err)
	}

	kva := mapf(filename, string(content))

	encoders := make([]*json.Encoder, nReduce)
	files := make([]*os.File, nReduce)
	for i := 0; i < nReduce; i++ {
		oname := reduceName(taskNumber, i)
		f, err := os.Create(oname)
		if err != nil {
			log.Fatalf("doMapTask: cannot create file %v: %v", oname, err)
		}
		files[i] = f
		encoders[i] = json.NewEncoder(f)
	}

	for _, kv := range kva {
		r := ihash(kv.Key) % nReduce
		err := encoders[r].Encode(&kv)
		if err != nil {
			log.Fatalf("doMapTask: error encoding kv %v: %v", kv, err)
		}
	}

	for i := 0; i < nReduce; i++ {
		files[i].Close()
	}
}

func doReduceTask(taskNumber int, nMap int, reducef func(string, []string) string) {
	var kva []KeyValue

	for i := 0; i < nMap; i++ {
		filename := reduceName(i, taskNumber)
		file, err := os.Open(filename)
		if err != nil {
			continue
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv)
		}
		file.Close()
	}

	sort.Slice(kva, func(i, j int) bool {
		return kva[i].Key < kva[j].Key
	})

	oname := "mr-out-" + strconv.Itoa(taskNumber)
	ofile, err := os.Create(oname)
	if err != nil {
		log.Fatalf("doReduceTask: cannot create output file %v: %v", oname, err)
	}

	i := 0
	for i < len(kva) {
		j := i + 1
		var values []string
		values = append(values, kva[i].Value)
		for j < len(kva) && kva[j].Key == kva[i].Key {
			values = append(values, kva[j].Value)
			j++
		}
		output := reducef(kva[i].Key, values)

		_, err := fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)
		if err != nil {
			log.Fatalf("doReduceTask: writing to file error: %v", err)
		}
		i = j
	}

	ofile.Close()
}

func reduceName(mapTask int, reduceTask int) string {
	return "mr-" + strconv.Itoa(mapTask) + "-" + strconv.Itoa(reduceTask) + ".json"
}

func reportTask(taskType TaskType, taskNumber int, workerID int) {
	args := ReportTaskArgs{
		TaskType:   taskType,
		TaskNumber: taskNumber,
		WorkerID:   workerID,
	}
	var reply ReportTaskReply
	ok := call("Coordinator.ReportTask", &args, &reply)
	if !ok {
		log.Printf("reportTask: RPC call ReportTask failed for task %v %d", taskType, taskNumber)
	}
}
