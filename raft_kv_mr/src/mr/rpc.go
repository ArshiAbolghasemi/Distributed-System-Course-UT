package mr

import (
	"os"
	"strconv"
)

type TaskType string

const (
	TaskTypeMap    TaskType = "Map"
	TaskTypeReduce TaskType = "Reduce"
	TaskTypeWait   TaskType = "Wait"
	TaskTypeExit   TaskType = "Exit"
)

type TaskArgs struct {
	WorkerID int
}

type TaskReply struct {
	TaskType   TaskType
	FileName   string
	TaskNumber int
	NReduce    int
	NMap       int
}

type ReportTaskArgs struct {
	TaskType   TaskType
	TaskNumber int
	WorkerID   int
}

type ReportTaskReply struct {
	Success bool
}

// coordinatorSock generates a unique-ish UNIX-domain socket name in /var/tmp,
// based on the user ID. We can't use the current directory because Athena AFS doesn't
// support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

// // ExampleArgs is an example structure for an RPC argument.
// type ExampleArgs struct {
// 	X int
// }

// // ExampleReply is an example structure for an RPC reply.
// type ExampleReply struct {
// 	Y int
// }
