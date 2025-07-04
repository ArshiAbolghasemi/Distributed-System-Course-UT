# KVMessaging - Concurrent Key-Value Messaging System

A networked key-value messaging system implemented in Go with a focus on concurrency and scalability.

## Overview

KVMessaging is a distributed key-value storage and messaging system that allows multiple clients to connect to a central server and perform operations on a shared data store. The system supports storing multiple values per key, real-time notifications for data changes, and concurrent access through a clean command-line interface.

### Code Implementation

The project consists of several key components:

#### 1. Key-Value Store (Core Data Structure)
- **KVStore Interface**: Defines operations for interacting with the key-value store
  ```go
  type KVStore interface {
      Put(key string, value []byte)
      Get(key string) []([]byte)
      Delete(key string)
      Update(key string, oldVal []byte, newVal []byte)
  }
  ```
- **Implementation**: Uses a map as the underlying storage where each key can have multiple values
  ```go
  type impl struct {
      storage map[string][]([]byte)
  }
  ```
- **Utility Functions**: Helper functions for finding and replacing elements in slices:
  ```go
  func FindIndex(s [][]byte, target []byte) int {
      for i, val := range s {
          if bytes.Equal(val, target) {
              return i
          }
      }
      return -1
  }
  
  func ReplaceElement[T any](s []T, i int, v T) []T {
      if i < 0 || i >= len(s) {
          return s
      }
      s = slices.Delete(s, i, i+1)
      return append(s[:i], append([]T{v}, s[i:]...)...)
  }
  ```

#### 2. Server Implementation
- **KeyValueServer Interface**: Defines methods for managing the server lifecycle
  ```go
  type KeyValueServer interface {
      Start(port int) error
      CountActive() int
      CountDropped() int
      Close()
  }
  ```
- **KVServer Implementation**: 
  - Structure:
    ```go
    type KVServer struct {
        listener           net.Listener
        port               int
        kvstore            kvstore.KVStore
        started            bool
        closed             bool
        clients            map[int]*Client
        nextClientID       int
        activeCount        int
        droppedCount       int
        opsChan            chan Command
        registerClientChan chan *Client
        removeClientChan   chan int
        countChan          chan chan int
        droppedChan        chan chan int
        shutdown           chan struct{}
    }
    ```
  - Features:
    - Listens for client connections on a specified port
    - Manages multiple concurrent client connections
    - Processes commands from clients to interact with the key-value store
    - Tracks active and dropped client connections
    - Uses channels for thread-safe communication between components
    - Notifies all clients when data changes

#### 3. Client Implementation
- **KVClient Interface**: Defines methods for client-server communication
  ```go
  type KVClient interface {
      Connect(address string) error
      SendMessage(message string) error
      ReadResponse() (string, error)
      Close()
  }
  ```
- **KVC Implementation**: 
  ```go
  type KVC struct {
      conn   net.Conn
      reader *bufio.Reader
      done   chan struct{}
  }
  ```
  - Establishes connection to the server using the configured protocol
  - Sends command messages to the server
  - Reads responses from the server
  - Provides clean connection lifecycle management

#### 4. Command Processing
- **Command Structure**: Represents operations on the key-value store
  ```go
  type Command struct {
      operation OperationType
      key       string
      value     []byte
      oldValue  []byte
      respChan  chan []byte
  }
  ```
- **Command Parsing**: Converts raw input from clients into structured commands
- **Protocol Format**: Uses colon-separated format (e.g., "PUT:key:value")
- **Operations**: Supports PUT, GET, DELETE, and UPDATE operations

#### 5. Runner Applications
- **Client Runner (crunner)**: 
  - Interactive command-line interface for connecting to the server
  - Provides color-coded help and command formatting
  - Handles user commands and server responses concurrently
  - Supports graceful termination

- **Server Runner (srunner)**:
  - Launches the server with configuration from the config package
  - Handles termination signals
  - Reports server statistics on shutdown

### Concurrency Model

The project uses Go's concurrency primitives to handle multiple concurrent connections and operations:

1. **Server-side Concurrency**:
   - **Multiple Goroutines**: Different aspects of server operation run in separate goroutines
     ```go
     go s.handleKVStore()      // Process key-value operations
     go s.manageClients()      // Manage client registry
     go s.acceptConnections()  // Accept new connections
     ```
   - **Per-Client Goroutines**: Each client has dedicated goroutines for reading, writing, and command processing
     ```go
     go s.handleClientRead(client)
     go s.handleClientWrite(client)
     go s.processClientCommands(client)
     ```

2. **Channel-based Communication**:
   - **Command Flow**: Commands pass through channels between components
     ```go
     s.opsChan <- cmd          // Send command to operation processor
     resp := <-cmd.respChan    // Receive response
     ```
   - **Client Management**: Registration and removal via channels
     ```go
     s.registerClientChan <- client  // Register new client
     s.removeClientChan <- clientID  // Remove disconnected client
     ```
   - **Coordination**: Channels for synchronizing operations
     ```go
     client.inChan  = make(chan Command)    // Commands from client
     client.outChan = make(chan []byte)     // Responses to client
     client.done    = make(chan struct{})   // Termination signal
     ```

3. **Thread-safe Access**:
   - All access to shared state is mediated through channels
   - No direct manipulation of shared data structures
   - Avoids locks and mutexes by using message passing

4. **Client Concurrency**:
   - Separate goroutines for handling user input and server responses
     ```go
     go handleRead(kvClient, stopChan)
     go handleClientCommands(kvClient, stopChan)
     ```
   - Coordinated shutdown via signal channels

5. **Notification System**:
   - Fan-out pattern for broadcasting changes to all connected clients
     ```go
     func (s *KVServer) notifyClients(notification string) {
         for id, client := range s.clients {
             select {
             case client.outChan <- []byte(notification + "\n"):
             default:
                 fmt.Printf("Failed to notify msg %s to client %d: buffer full\n", notification, id)
             }
         }
     }
     ```
   - Non-blocking sends to prevent slow clients from affecting others

## Project Structure

```
kvmsg/
├── config/                 # Configuration management
├── internal/
│   ├── kvstore/            # Key-value store implementation
│   │   ├── kvstore_api.go      # Interface definition
│   │   └── kvstore_impl.go         # Implementation
│   ├── client/             # Client implementation
│   │   ├── client_api.go       # Interface definition
│   │   └── client_impl.go         # Implementation  
│   ├── server/             # Server implementation
│   │   ├── server_api.go       # Interface definition
│   │   ├── server_impl.go         # Implementation
│   │   └── command.go      # Command processing
│   └── util/               # Utility functions
├── cmd/
│   ├── clien/            # Client runner
│   │   └── crunner.go         # Client entry point
│   └── serverr/            # Server runner
│       └── srunner.go         # Server entry point
└── README.md               # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.18 or higher

### Building the Project

### Running the Server

```bash
make server
```

### Running the Client

```bash
make client
```

## Usage

### Available Commands

- **GET**: Retrieve values for a key
  ```
  GET:key
  ```

- **PUT**: Store a new value for a key
  ```
  PUT:key:value
  ```

- **DELETE**: Remove a key and all its values
  ```
  DELETE:key
  ```

- **UPDATE**: Replace a specific value for a key
  ```
  UPDATE:key:oldValue:newValue
  ```

### Example Session

```
Connected to KV server at localhost:8080

Available operations:
  GET     - Retrieve a value
  PUT     - Store a new key-value pair
  UPDATE  - Modify an existing value
  DELETE  - Remove a key

Command formats:
  Get:key
  Put:key:value
  Update:key:oldValue:newValue
  Delete:key

> PUT:user:john
Server: OK

> PUT:user:alice
Server: OK

> GET:user
Server: john
alice

> UPDATE:user:john:johnsmith
Server: OK

> GET:user
Server: johnsmith
alice

> DELETE:user
Server: OK

> GET:user
Server: NOT_FOUND
```

## Technical Details

### Multi-Value Support

Unlike traditional key-value stores where each key maps to a single value, this implementation allows storing multiple values per key. When using PUT, values are appended to the existing list for that key.

### Real-time Notifications

Clients receive notifications when other clients modify the data store:
```go
// Server notifies all clients of changes
s.notifyClients(fmt.Sprintf("NOTIFY PUT %s", cmd.key))
s.notifyClients(fmt.Sprintf("NOTIFY DELETE %s", cmd.key))
s.notifyClients(fmt.Sprintf("NOTIFY UPDATE %s", cmd.key))
```

This allows for real-time synchronization between clients.

### Connection Management

The server tracks active connections and gracefully handles disconnects:
```go
// Track connection statistics
case client := <-s.registerClientChan:
    s.clients[client.id] = client
    s.activeCount++

case clientID := <-s.removeClientChan:
    if client, exists := s.clients[clientID]; exists {
        delete(s.clients, clientID)
        s.activeCount--
        s.droppedCount++
        // Clean up resources...
    }
```

It maintains statistics about active and dropped connections that can be queried:
```go
// Query connection statistics
func (s *KVServer) CountActive() int {
    respChan := make(chan int)
    s.countChan <- respChan
    return <-respChan
}

func (s *KVServer) CountDropped() int {
    respChan := make(chan int)
    s.droppedChan <- respChan
    return <-respChan
}
```

### Concurrency Implementation Details

The project's concurrency model is designed to maximize throughput while ensuring data consistency:

1. **Command Processing Pipeline**:
   - Client request → Command parsing → Operation queue → KV store execution → Response
   - Each step runs in a separate goroutine, allowing for parallel processing

2. **Non-blocking Design**:
   - All operations use buffered channels where appropriate
   - Server never blocks on client operations
   - Clients with full buffers are handled gracefully (dropped messages are logged)

3. **Graceful Shutdown**:
   - Signal handling for clean termination
   ```go
   signalChan := make(chan os.Signal, 1)
   signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
   // ...
   kvServer.Close()
   ```
   - Resource cleanup on exit
   - Connection statistics reporting

4. **Fan-out Pattern for Notifications**:
   - Changes to the data store trigger notifications to all connected clients
   - Non-blocking sends to prevent slow clients from affecting others

This architecture allows the system to handle many concurrent connections while maintaining data consistency and responsive user experience, showcasing Go's powerful concurrency primitives and channel-based communication pattern.
