package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/eserafini/golangtechweek/pkg/workerpool"
)

type NumberJob struct {
	Number int
}

type NumberResult struct {
	Value int
	WorkerID int
	Timestamp time.Time
}

func proccessNumber(ctx context.Context, job workerpool.Job) workerpool.Result {
	number := job.(NumberJob).Number
	workerID := number % 3

	sleepTime := time.Duration(800 + rand.Intn(400)) * time.Millisecond
	time.Sleep(sleepTime)

	return NumberResult{
		Value: number,
		WorkerID: workerID,
		Timestamp: time.Now(),
	}
}

func main() {
	maxValue := 20
	bufferSize := 10

	pool := workerpool.New(proccessNumber, workerpool.Config{
		WorkerCount: 3,
	})

	inputCh := make(chan workerpool.Job, bufferSize)
	ctx := context.Background()

	resultCh, err := pool.Start(ctx, inputCh)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(maxValue)

	fmt.Println("Initializing pool with max ", maxValue)

	go func() {
		for i := 0; i < maxValue; i++ {
			inputCh <- NumberJob{Number: i}
		}
		close(inputCh)
	}()

	go func() {
		for result := range resultCh {
			r := result.(NumberResult)
			fmt.Printf("Number: %d", r.Value)
			wg.Done()
		}
	}()

	wg.Wait()
}
