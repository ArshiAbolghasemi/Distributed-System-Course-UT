package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type TaskStatus struct {
	FileName   string
	TaskNumber int
	Done       bool
	WorkerID   int
	StartTime  time.Time
}

type Coordinator struct {
	mu           sync.Mutex
	mapTasks     []TaskStatus
	reduceTasks  []TaskStatus
	nReduce      int
	nMap         int
	mapPhaseDone bool
}

func (c *Coordinator) GetTask(args *TaskArgs, reply *TaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.mapPhaseDone {
		for i := range c.mapTasks {
			if !c.mapTasks[i].Done && (c.mapTasks[i].WorkerID == 0 || time.Since(c.mapTasks[i].StartTime) > 10*time.Second) {
				c.mapTasks[i].WorkerID = args.WorkerID
				c.mapTasks[i].StartTime = time.Now()
				reply.TaskType = TaskTypeMap
				reply.FileName = c.mapTasks[i].FileName
				reply.TaskNumber = c.mapTasks[i].TaskNumber
				reply.NReduce = c.nReduce
				reply.NMap = c.nMap
				log.Printf("Assigned Map task: %v to worker %d\n", c.mapTasks[i].TaskNumber, args.WorkerID)
				return nil
			}
		}
		reply.TaskType = TaskTypeWait
		return nil
	}

	for i := range c.reduceTasks {
		if !c.reduceTasks[i].Done && (c.reduceTasks[i].WorkerID == 0 || time.Since(c.reduceTasks[i].StartTime) > 10*time.Second) {
			c.reduceTasks[i].WorkerID = args.WorkerID
			c.reduceTasks[i].StartTime = time.Now()
			reply.TaskType = TaskTypeReduce
			reply.TaskNumber = c.reduceTasks[i].TaskNumber
			reply.NReduce = c.nReduce
			reply.NMap = c.nMap
			log.Printf("Assigned Reduce task: %v to worker %d\n", c.reduceTasks[i].TaskNumber, args.WorkerID)
			return nil
		}
	}
	allDone := true
	for i := range c.reduceTasks {
		if !c.reduceTasks[i].Done {
			allDone = false
			break
		}
	}
	if allDone {
		reply.TaskType = TaskTypeExit
	} else {
		reply.TaskType = TaskTypeWait
	}
	return nil
}

func (c *Coordinator) ReportTask(args *ReportTaskArgs, reply *ReportTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if args.TaskType == TaskTypeMap {
		if args.TaskNumber < len(c.mapTasks) {
			c.mapTasks[args.TaskNumber].Done = true
			log.Printf("Map task %d completed by worker %d\n", args.TaskNumber, args.WorkerID)
			allDone := true
			for _, task := range c.mapTasks {
				if !task.Done {
					allDone = false
					break
				}
			}
			if allDone {
				c.mapPhaseDone = true
				log.Printf("All Map tasks completed. Moving to Reduce phase.\n")
			}
		}
	} else if args.TaskType == TaskTypeReduce {
		if args.TaskNumber < len(c.reduceTasks) {
			c.reduceTasks[args.TaskNumber].Done = true
			log.Printf("Reduce task %d completed by worker %d\n", args.TaskNumber, args.WorkerID)
		}
	}
	reply.Success = true
	return nil
}

func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.mapPhaseDone {
		return false
	}
	for _, task := range c.reduceTasks {
		if !task.Done {
			return false
		}
	}
	return true
}

func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		nReduce:      nReduce,
		nMap:         len(files),
		mapPhaseDone: false,
	}
	for idx, file := range files {
		c.mapTasks = append(c.mapTasks, TaskStatus{
			FileName:   file,
			TaskNumber: idx,
			Done:       false,
		})
	}

	for i := 0; i < nReduce; i++ {
		c.reduceTasks = append(c.reduceTasks, TaskStatus{
			TaskNumber: i,
			Done:       false,
		})
	}

	c.server()
	return &c
}
