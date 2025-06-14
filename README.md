# Go WorkerPool

A simple and efficient worker pool implementation in Go that allows you to process jobs concurrently with a configurable number of workers.

## Features

- Configurable number of workers
- Type-safe job and result handling
- Context support for cancellation
- Graceful shutdown
- Built-in logging support
- Thread-safe state management

## Installation

```bash
go get github.com/eserafini/go-workerpool
```

## Usage

Here's a simple example of how to use the worker pool to process numbers:

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/eserafini/go-workerpool/pkg/workerpool"
)

// Define your job type
type NumberJob struct {
    Number int
}

// Define your result type
type NumberResult struct {
    Value     int
    WorkerID  int
    Timestamp time.Time
}

// Define your processing function
func processNumber(ctx context.Context, job workerpool.Job) workerpool.Result {
    number := job.(NumberJob).Number
    workerID := number % 3

    // Simulate some work
    time.Sleep(100 * time.Millisecond)

    return NumberResult{
        Value:     number,
        WorkerID:  workerID,
        Timestamp: time.Now(),
    }
}

func main() {
    // Create a new worker pool with 3 workers
    pool := workerpool.New(processNumber, workerpool.Config{
        WorkerCount: 3,
    })

    // Create input channel with buffer size
    inputCh := make(chan workerpool.Job, 10)
    ctx := context.Background()

    // Start the worker pool
    resultCh, err := pool.Start(ctx, inputCh)
    if err != nil {
        panic(err)
    }

    // Create a wait group to wait for all results
    var wg sync.WaitGroup
    wg.Add(20) // Number of jobs to process

    // Send jobs to the pool
    go func() {
        for i := 0; i < 20; i++ {
            inputCh <- NumberJob{Number: i}
        }
        close(inputCh)
    }()

    // Process results
    go func() {
        for result := range resultCh {
            r := result.(NumberResult)
            fmt.Printf("Processed number %d by worker %d at %v\n", 
                r.Value, r.WorkerID, r.Timestamp)
            wg.Done()
        }
    }()

    // Wait for all results
    wg.Wait()
}
```

## Configuration

The worker pool can be configured using the `Config` struct:

```go
type Config struct {
    WorkerCount int           // Number of workers in the pool
    Logger      *slog.Logger  // Optional logger
}
```

## API

### Types

- `Job interface{}` - Interface for job types
- `Result interface{}` - Interface for result types
- `ProccessFun func(ctx context.Context, job Job) Result` - Function type for processing jobs

### Methods

- `New(proccessFunc ProccessFun, config Config) workerPool` - Creates a new worker pool
- `Start(ctx context.Context, inputChan <-chan Job) (<-chan Result, error)` - Starts the worker pool
- `Stop() error` - Stops the worker pool gracefully
- `IsRunning() bool` - Checks if the worker pool is running

## License

MIT License 