# Key-Value Server

A distributed key-value storage system with versioned updates and distributed locking capabilities, built in Go.

## Overview

This project implements a fault-tolerant key-value server that supports versioned operations and distributed locking. The system is designed to handle network failures gracefully while maintaining data consistency through optimistic concurrency control.

## Architecture

The system consists of three main components:

### 1. KV Server (`server.go`)
The core storage engine that manages key-value pairs with versioning.

**Key Features:**
- **Versioned Storage**: Each key has an associated version number that increments with each update
- **Optimistic Concurrency Control**: Updates only succeed if the provided version matches the current version
- **Thread-Safe Operations**: Uses mutex locks to ensure concurrent access safety
- **Conditional Puts**: Supports creating new keys (version 0) or updating existing keys with version matching

**API:**
- `Get(key)` → Returns value, version, and error status
- `Put(key, value, version)` → Updates key if version matches, returns error status

### 2. Client/Clerk (`client.go`)
The client-side interface that handles communication with the KV server and implements retry logic for fault tolerance.

**Retry Logic:**
- **Get Operations**: Retry indefinitely until successful
- **Put Operations**:
  - First failure → Continue retrying
  - Subsequent `ErrVersion` after network failure → Return `ErrMaybe` (operation may have succeeded)

### 3. Distributed Lock (`lock.go`)
A distributed locking mechanism built on top of the key-value store using test-and-set semantics.

**Key Features:**
- **Test-and-Set Locking**: Atomically checks and sets lock state
- **Client Identification**: Each lock instance has a unique client ID
- **Deadlock Prevention**: Clients can re-acquire locks they already hold
- **Initialization**: Ensures lock keys are properly initialized on the server

**Lock States:**
- `__unlocked`: Lock is available
- `{clientId}`: Lock is held by the specified client

## Data Flow

```
Client Application
       ↓
    Clerk (client.go)
       ↓ RPC Calls
   KVServer (server.go)
       ↓
  keyValueStore (in-memory)
```

## Usage Examples

### Basic Key-Value Operations

```go
// Create a clerk (client)
clerk := MakeClerk(clnt, serverAddress)

// Get a value
value, version, err := clerk.Get("mykey")
if err == rpc.ErrNoKey {
    // Key doesn't exist
}

// Put a new value (version 0 for new keys)
err = clerk.Put("mykey", "myvalue", 0)

// Update existing value
err = clerk.Put("mykey", "newvalue", version)
if err == rpc.ErrVersion {
    // Version mismatch - someone else updated the key
}
```

### Distributed Locking

```go
// Create a lock
lock := MakeLock(clerk, "mylockkey")

// Acquire the lock
lock.Acquire() // Blocks until lock is acquired

// Critical section
// ... do work ...

// Release the lock
lock.Release()
```

## Error Handling Strategy

The system implements a error handling strategy:

1. **Network Failures**: Automatically retry with exponential backoff
2. **Version Conflicts**: Return `ErrVersion` for immediate conflicts
3. **Ambiguous States**: Return `ErrMaybe` when operation outcome is uncertain
4. **Missing Keys**: Return `ErrNoKey` for non-existent keys

## Concurrency Model

- **Server-Side**: Uses mutex locks for thread-safe access to the key-value store
- **Client-Side**: Each clerk instance can be used concurrently by multiple goroutines
- **Versioning**: Prevents lost updates through optimistic concurrency control


## Testing

The system is designed to work with the provided test framework:
- Supports network partition simulation
- Handles message loss and reordering
- Validates correctness under concurrent access
