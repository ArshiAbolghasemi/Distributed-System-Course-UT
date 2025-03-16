# Go Squarer Package Documentation

## Overview

The `squarer` package provides a concurrent service for squaring integers. It uses Go's concurrency primitives (goroutines and channels) to process integers asynchronously. This design allows for concurrent processing while ensuring each value is handled correctly.

## Package Components

### Core Files
- `squarer_api.go` - Contains the interface definition
- `squarer_impl.go` - Contains the implementation
- `squarer_test.go` - Contains tests

## API

### SquarerImpl Structure

The core structure that implements the squaring service:

```go
type SquarerImpl struct {
    input  chan int    // Channel to receive input numbers
    output chan int    // Channel to send squared results
    close  chan bool   // Channel to signal shutdown
    closed chan bool   // Channel to confirm shutdown completed
}
```

### Methods

#### Initialize

```go
func (sq *SquarerImpl) Initialize(input chan int) <-chan int
```

- **Purpose**: Sets up the channels and starts the worker goroutine
- **Parameters**: `input` - A channel from which integers will be read
- **Returns**: A channel from which squared results can be read
- **Behavior**: 
  - Stores the input channel reference
  - Creates new channels for output, close signals, and close confirmation
  - Starts the worker goroutine
  - Returns the output channel for clients to receive results

#### Close

```go
func (sq *SquarerImpl) Close()
```

- **Purpose**: Shuts down the worker goroutine cleanly
- **Behavior**:
  - Sends a signal on the close channel
  - Waits for confirmation on the closed channel
  - Returns after worker confirms shutdown

#### work (internal method)

```go
func (sq *SquarerImpl) work()
```

- **Purpose**: Processes integers from the input channel, squares them, and sends results to the output channel
- **Behavior**:
  - Implements a state machine with two phases:
    1. Receive a value and square it
    2. Send the squared value to output
  - Uses channel reassignment to control flow between phases
  - Responds to close signals for clean shutdown
  - Runs in its own goroutine

## Implementation Details

### State Machine

The worker uses a clever state machine pattern with a `select` statement:

```go
func (sq *SquarerImpl) work() {
    var toPush int     // Value to be sent to output
    dummy := make(chan int)  // Never used channel for state management
    pushOn := dummy    // Channel to push to (initially dummy)
    pullOn := sq.input // Channel to pull from (initially input)
    
    for {
        select {
        case unsquared := <-pullOn:
            toPush = unsquared * unsquared  // Square the input
            pushOn = sq.output  // Set to push to real output
            pullOn = nil        // Don't pull until current value is pushed
            
        case pushOn <- toPush:
            pushOn = dummy      // Reset to dummy (won't send anything)
            pullOn = sq.input   // Ready to pull next input
            
        case <-sq.close:
            sq.closed <- true   // Confirm shutdown
            return              // Exit goroutine
        }
    }
}
```

This design ensures:
- Each value is fully processed before the next is handled
- The goroutine alternates between reading and writing
- No race conditions occur
- Clean shutdown is possible

## Testing

The package includes several tests to verify functionality:

### TestBasicCorrectness

Tests basic functionality by:
- Sending a single value (2)
- Verifying the correct result is returned (4)

### TestSequentialNumbers

Tests handling of multiple sequential values by:
- Sending a sequence of numbers (1, 3, 5, 10)
- Verifying each result in order (1, 9, 25, 100)

### TestConcurrentProcessing

Tests handling of concurrent operations by:
- Creating multiple goroutines to send values concurrently
- Collecting results through a buffer channel
- Verifying all expected squared values are received

## Running Tests

To run tests:

```bash
# With Go modules
cd /path/to/squarer
go test -v squarer_test.go squarer_impl.go squarer_api.go
