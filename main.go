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
