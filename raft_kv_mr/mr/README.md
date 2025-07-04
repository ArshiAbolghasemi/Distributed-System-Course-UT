# MapReduce Distributed System

### Overview

This project implements a distributed MapReduce system consisting of two main components:
- **Coordinator**: Manages and distributes tasks
- **Workers**: Execute the assigned tasks

In this system, one Coordinator process is initialized first, followed by one or more Worker processes running in parallel. While in a real-world scenario these Workers would run on separate machines, for simplicity in this project, all processes run on a single machine.

## Architecture

### Communication
- Workers communicate with the Coordinator using RPC (Remote Procedure Call) mechanisms
- The system handles task assignment, execution, and failure recovery

### Worker Process Flow
1. Worker requests a task (Map or Reduce) from the Coordinator
2. Worker reads input for the task from one or more files
3. Worker executes the assigned task
4. Worker writes output to one or more files
5. Worker requests a new task (loop continues)

### Fault Tolerance
- The Coordinator detects and handles Worker failures
- If a Worker fails to complete a task within a reasonable time frame (10 seconds in this project), the Coordinator reassigns the task to another Worker

## Components

### Coordinator
- Maintains the state of all tasks (pending, in-progress, completed)
- Assigns tasks to Workers upon request
- Monitors task execution time and handles timeouts
- Tracks overall job progress

### Worker
- Runs in a continuous loop requesting tasks
- Processes Map and Reduce tasks as assigned
- Handles input and output file operations
- Reports task completion to the Coordinator

## Implementation Details

### mrsequential.go
This file implements a sequential (non-distributed) version of MapReduce for testing and comparison:
- Loads Map/Reduce functions from a plugin
- Processes all input files through the Map function
- Collects and sorts all intermediate key-value pairs
- Applies the Reduce function to each unique key
- Outputs results to a single file

### mrworker.go
Entry point for worker processes:
- Loads Map/Reduce functions from a plugin
- Initiates the Worker process

### mrcoordinator.go
Entry point for the coordinator process:
- Creates a coordinator with input files and reduce tasks
- Monitors until all work is complete

### worker.go
Implements the worker's main loop and task processing:
- Continuously requests tasks via RPC
- Processes Map tasks by reading input, applying the Map function, and partitioning output
- Processes Reduce tasks by collecting data, sorting by key, and applying the Reduce function
- Reports task completion back to the coordinator

### rpc.go
Defines data structures for RPC communication:
- Task types (Map, Reduce, Wait, Exit)
- Request/response structures for task assignment and completion reporting

### coordinator.go
Implements the coordinator that manages task distribution:
- Tracks the status of all tasks
- Assigns tasks to workers, prioritizing Map before Reduce
- Handles failures by reassigning tasks that take too long
- Determines when all work is complete

The implementation focuses on:
- Robust RPC-based communication between Coordinator and Workers
- Efficient task distribution and load balancing
- Proper handling of task failures and timeouts
- Correct implementation of the MapReduce paradigm

## Getting Started

### Available Applications
The system includes several test applications:
- `wc`: Word count
- `indexer`: Indexing
- `crash`: Tests fault tolerance with deliberate crashes
- `nocrash`: Tests normal operation without crashes
- `rtiming`: Tests timing
- `mtiming`: Tests timing for map tasks
- `jobcount`: Counts jobs
- `early_exit`: Tests early exit behavior

### Building and Running

The project includes a Makefile with several useful commands:

#### Building Application Plugins
```
make wc       # Builds the word count plugin
make indexer  # Builds the indexer plugin
# etc. for other applications
```

#### Running in Sequential Mode
```
make run-sequential APP=wc  # Run word count in sequential mode
```

#### Running in Distributed Mode
Start the coordinator:
```
make run-coordinator
```

Then start one or more workers in separate terminals:
```
make run-worker APP=wc
```

#### Testing
```
make test  # Runs all tests
```

#### Cleaning Up
```
make clean  # Removes temporary files and build artifacts
```

#### Help
```
make help  # Shows available make commands
```

## Requirements

- Go programming environment
