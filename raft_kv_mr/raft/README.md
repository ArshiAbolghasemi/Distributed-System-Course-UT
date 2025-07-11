# Raft Consensus Algorithm Implementation

A comprehensive implementation of the Raft consensus algorithm in Go, providing distributed consensus and fault tolerance for distributed systems. This implementation follows the Raft consensus algorithm as described in the original paper "In Search of an Understandable Consensus Algorithm" by Diego Ongaro and John Ousterhout.

## Overview

Raft is designed to be more understandable than other consensus algorithms like Paxos while providing equivalent fault tolerance and performance. The Raft algorithm is a consensus protocol designed to be easy to understand and implement, providing fault-tolerant distributed state machine replication.

## Features

- **Leader Election**: Automatic leader election with randomized timeouts to prevent split votes
- **Log Replication**: Reliable log replication across cluster nodes with consistency guarantees
- **Safety**: Ensures consistency even during network partitions and node failures
- **Persistence**: Durable storage of critical state information with automatic recovery
- **Log Compaction**: Efficient snapshotting to prevent unbounded log growth
- **Fast Backup**: Efficient conflict resolution optimization in log replication
- **Network Resilience**: Handles network partitions and message loss gracefully
- **Concurrent Safety**: Thread-safe operations with proper synchronization mechanisms

## Architecture

### Key Components

The implementation consists of four main components that handle different aspects of the Raft protocol:

#### 1. Log Management (`raft_log.go`)

**Purpose**: Manages the Raft log entries and provides utilities for log operations.

**Key Features**:
- **Log Entry Types**: Supports both client commands and no-op entries
- **Log Search**: Binary search functions to find entries by term or command index
- **Indexing**: Maintains both logical command indexes and actual log indexes

**Core Types**:
```go
type raftLogEntry struct {
    Command      any          // State machine command
    CommandIndex int          // Index in state machine (ignoring no-op entries)
    LogIndex     int          // Real log index without compaction
    Term         int          // Raft term number
    LogEntryType logEntryType // Entry type (no-op or client command)
}
```

**Key Functions**:
- `findFirstIndexWithTerm(term)` - Find first log entry with given term
- `findLastIndexWithTerm(term)` - Find last log entry with given term
- `findLastCommandIndex(commandIndex)` - Find log entry by command index

#### 2. Leader Election (`raft_election.go`)

**Purpose**: Implements the Raft leader election protocol to ensure exactly one leader per term.

**Key Features**:
- **Election Timeout**: Randomized timeouts to prevent split votes
- **Vote Collection**: Concurrent vote gathering from all peers
- **Term Management**: Handles term transitions and vote validation
- **State Transitions**: Manages candidate → leader or candidate → follower transitions

**Election Process**:
1. **Election Trigger**: Starts when election timeout expires
2. **Vote Request**: Sends `RequestVote` RPCs to all peers
3. **Vote Collection**: Collects votes concurrently with timeout handling
4. **Leadership**: Becomes leader if majority votes received

**Key Functions**:
- `elect()` - Initiates election process
- `shouldStartElection()` - Determines if election should start
- `sendVotes()` - Sends vote requests to all peers
- `collectVotes()` - Collects and processes vote responses

#### 3. Log Replication (`raft_replication.go`)

**Purpose**: Handles log synchronization between leader and followers to maintain consistency.

**Key Features**:
- **AppendEntries**: Replicates log entries to followers
- **Snapshot Installation**: Handles log compaction via snapshots
- **Conflict Resolution**: Implements fast log backup for inconsistencies
- **Commit Index Management**: Updates commit index based on majority replication

**Replication Process**:
1. **Replication Loop**: Continuous replication to each follower
2. **Entry Synchronization**: Sends new entries or handles conflicts
3. **Snapshot Fallback**: Uses snapshots when logs are too far behind
4. **Commit Advancement**: Updates commit index when majority replicated

**Key Functions**:
- `replicate(server)` - Main replication loop for a peer
- `replicateAppendEntries()` - Handles log entry replication
- `replicateInstallSnapshot()` - Handles snapshot installation
- `setCommitIndexAsMajorityReplicatedIndex()` - Updates commit index

#### 4. State Machine Application (`raft_apply.go`)

**Purpose**: Applies committed log entries to the state machine in order.

**Key Features**:
- **Ordered Application**: Ensures logs are applied in correct order
- **Snapshot Application**: Handles state machine snapshots
- **Client Communication**: Sends applied commands to clients via channels
- **Graceful Shutdown**: Handles server shutdown cleanly

**Application Process**:
1. **Commit Detection**: Monitors for commit index changes
2. **Log Application**: Applies newly committed entries to state machine
3. **Client Notification**: Notifies clients of applied commands
4. **Snapshot Handling**: Applies snapshots when received

**Key Functions**:
- `applier()` - Main application loop
- `applyLog()` - Applies committed log entries
- `applySnapshot()` - Applies received snapshots

#### 5. Raft Server (`raft.go`)
The main Raft implementation containing:
- **State Management**: Follower, Candidate, and Leader states
- **RPC Handlers**: RequestVote, AppendEntries, and InstallSnapshot
- **Persistence**: Automatic state persistence and recovery
- **Timers**: Election and heartbeat timeout management

#### 6. Test Server (`server.go`)
A test framework that provides:
- **State Machine Simulation**: Simulates a key-value store
- **Snapshot Testing**: Validates snapshotting functionality
- **Consistency Checking**: Verifies log consistency across nodes

## State Machine

The Raft implementation maintains several types of state:

### Persistent State (survives restarts)
- `currentTerm`: Latest term server has seen
- `votedFor`: CandidateId that received vote in current term
- `log[]`: Log entries with commands and metadata
- `snapshot`: Compressed state for log compaction

### Volatile State (all servers)
- `commitIndex`: Index of highest log entry known to be committed
- `lastApplied`: Index of highest log entry applied to state machine

### Volatile State (leaders only)
- `nextIndex[]`: Index of next log entry to send to each server
- `matchIndex[]`: Index of highest log entry known to be replicated

## Usage

### Basic Setup

```go
import (
    "github.com/ArshiAbolghasemi/disgo/labrpc"
    "github.com/ArshiAbolghasemi/disgo/raftapi"
    "github.com/ArshiAbolghasemi/disgo/tester1"
)

// Create peer connections
peers := make([]*labrpc.ClientEnd, 3)
// ... initialize peer connections ...

// Create persister for durable storage
persister := tester.MakePersister()

// Create apply channel for state machine
applyCh := make(chan raftapi.ApplyMsg)

// Create Raft instance
rf := Make(peers, 0, persister, applyCh)
```

### Server Initialization

```go
// Example server initialization
rf := &Raft{
    peers:     peers,
    persister: persister,
    me:        me,
    // ... other initialization
}

// Start the server components
go rf.ticker()        // Election timer
go rf.applier()       // State machine applier
// Replication started when becoming leader
```

### Starting Agreement on Commands

```go
// Submit a command to the Raft cluster
command := "SET key value"
index, term, isLeader := rf.Start(command)

if isLeader {
    fmt.Printf("Command submitted at index %d, term %d\n", index, term)
} else {
    fmt.Println("Not the leader, command rejected")
}
```

### Handling Applied Commands

```go
go func() {
    for msg := range applyCh {
        if msg.CommandValid {
            // Apply command to state machine
            fmt.Printf("Applying command at index %d: %v\n", 
                      msg.CommandIndex, msg.Command)
        } else if msg.SnapshotValid {
            // Handle snapshot
            fmt.Printf("Installing snapshot up to index %d\n", 
                      msg.SnapshotIndex)
        }
    }
}()
```

### Creating Snapshots

```go
// Create snapshot when log grows too large
snapshot := createSnapshot() // your snapshot creation logic
rf.Snapshot(lastAppliedIndex, snapshot)
```

## RPC Interface

### RequestVote RPC
Used during leader election to request votes from other servers.

```go
type RequestVoteArgs struct {
    Term         int // Candidate's term
    CandidateId  int // Candidate requesting vote
    LastLogIndex int // Index of candidate's last log entry
    LastLogTerm  int // Term of candidate's last log entry
}

type RequestVoteReply struct {
    Term        int  // Current term for candidate to update itself
    VoteGranted bool // True means candidate received vote
}
```

### AppendEntries RPC
Used for log replication and heartbeats.

```go
type AppendEntriesArgs struct {
    Term         int     // Leader's term
    LeaderId     int     // Leader ID for client redirection
    PrevLogIndex int     // Index of log entry immediately preceding new ones
    PrevLogTerm  int     // Term of prevLogIndex entry
    Entries      []Entry // Log entries to store (empty for heartbeat)
    LeaderCommit int     // Leader's commitIndex
}

type AppendEntriesReply struct {
    Term    int  // Current term for leader to update itself
    Success bool // True if follower contained entry matching prevLogIndex
    XTerm   int  // Term in the conflicting entry (optimization)
    XIndex  int  // Index of first entry with XTerm (optimization)
    XLen    int  // Log length (optimization)
}
```

### InstallSnapshot RPC
Used for log compaction and catch-up.

```go
type InstallSnapshotArgs struct {
    Term                     int    // Leader's term
    LeaderId                 int    // Leader ID
    LastIncludedCommandIndex int    // Snapshot command index
    LastIncludedLogIndex     int    // Snapshot log index
    LastIncludedTerm         int    // Snapshot term
    Data                     []byte // Snapshot data
}
```

## Thread Safety and Fault Tolerance

### Thread Safety
All components use proper locking mechanisms to ensure thread safety in a concurrent environment. The implementation uses:
- Mutex locks for critical sections
- Channels for inter-goroutine communication
- Atomic operations where appropriate

### Fault Tolerance
The implementation handles various failure scenarios:
- **Network Partitions**: Continues operation with majority
- **Server Failures**: Automatic leader election and log recovery
- **Message Loss**: Retry mechanisms with timeouts
- **Concurrent Operations**: Proper synchronization and state management

## Performance Characteristics

### Throughput
- Optimized for typical distributed system workloads
- Batching support for improved efficiency
- Parallel replication to followers

### Latency
- Single RTT for normal case operations
- Fast leader election with randomized timeouts
- Optimized log matching with conflict resolution

### Fault Tolerance
- Tolerates up to (n-1)/2 server failures in n-server cluster
- Handles network partitions gracefully
- Automatic recovery after partition healing

### Performance Optimizations
- **Parallel Processing**: Concurrent RPCs to multiple servers
- **Fast Backup**: Efficient conflict resolution in log replication
- **Batching**: Multiple log entries sent in single RPC
- **Snapshot Compaction**: Reduces log size for efficient storage

## Configuration

### Timeouts
- **Election Timeout**: 800-1300ms (randomized)
- **Heartbeat Timeout**: 100ms
- **Retry Timeout**: 200ms

### Snapshot Configuration
- **Snapshot Interval**: Every 10 committed entries
- Automatic triggering based on log growth
- Configurable via `SnapShotInterval` constant

## Testing

### Running Tests
```bash
# Run basic Raft tests
go test -run

# Run with race detection
go test -race -run

# Run leader election test
go test -run 3A

#Run Leader election test with race
go test -race -run 3A

# Run log test
go test -run 3B

#Run log test with race
go test -race -run 3B

# Run persister election test
go test -run 3C

#Run persister test with race
go test -race -run 3C

# Run snapshot test
go test -run 3D

#Run snapshot with race
go test -race -run 3D

```

## Logging and Debugging

Debug logging is provided through `DPrintf` with different topics:
- `tElection` - Election-related events
- `tVote` - Vote request/response events
- `tSendAppend` - Log replication events
- `tSendSnapshot` - Snapshot installation events
- `tApply` - State machine application events
- `tHeartbeat` - Heartbeat-related events
