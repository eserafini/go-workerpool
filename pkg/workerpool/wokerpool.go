package workerpool

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Job interface{}

type Result interface{}

type ProccessFun func(ctx context.Context, job Job) Result

type WorkerPool interface {
	Start(ctx context.Context, inputChan <-chan Job) (<-chan Result, error)
	Stop() error
	IsRunning() error
}
type State int

const (
	StateIdle State = iota
	StateRunning
	StateStopping
)

type Config struct {
	WorkerCount int
	Logger      *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		WorkerCount: 1,
		Logger:      slog.Default(),
	}
}

type workerPool struct {
	workerCount  int
	proccessFunc ProccessFun
	logger       *slog.Logger
	state        State
	stateMutex   sync.Mutex
	stopChan     chan struct{}
	stopWg       sync.WaitGroup
}

func New(proccessFunc ProccessFun, config Config) workerPool {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 1
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return workerPool{
		proccessFunc: proccessFunc,
		workerCount:  config.WorkerCount,
		stopChan:     make(chan struct{}),
		state:        StateIdle,
		logger:       config.Logger,
	}
}

func (wp *workerPool) Start(ctx context.Context, inputChan <-chan Job) (<-chan Result, error) {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock()

	if wp.state != StateIdle {
		return nil, fmt.Errorf("worker pool is not idle")
	}

	resultChan := make(chan Result)
	wp.state = StateRunning
	wp.stopChan = make(chan struct{})

	wp.stopWg.Add(wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		go wp.worker(ctx, i, inputChan, resultChan)
	}

	go func() {
		wp.stopWg.Wait()
		close(resultChan)

		wp.stateMutex.Lock()
		wp.state = StateIdle
		wp.stateMutex.Unlock()
	}()

	return resultChan, nil
}

func (wp *workerPool) Stop() error {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock()

	if wp.state != StateRunning {
		return fmt.Errorf("worker pool is not running")
	}

	wp.state = StateStopping
	close(wp.stopChan)

	wp.stopWg.Wait()

	wp.state = StateIdle
	return nil
}

func (wp *workerPool) IsRunning() bool {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock()
	return wp.state == StateRunning
}

func (wp *workerPool) worker(ctx context.Context, id int, inputChan <-chan Job, resultChan chan<- Result) {
	wp.logger.Info("worker started", "worker_id", id)

	for {
		select {
		case <-wp.stopChan:
			wp.logger.Info("worker stopped", "worker_id", id)
			return
		case <-ctx.Done():
			wp.logger.Info("context canclled, stopping worker", "worker_id", id)
		case job, ok := <-inputChan:
			if !ok {
				wp.logger.Info("Channel closed, stopping worker", "worker_id", id)
				return
			}

			result := wp.proccessFunc(ctx, job)
			select {
			case resultChan <- result:
			case <-wp.stopChan:
				wp.logger.Info("Woker stopped, result not sended", "worker_id", id)
				return
			case <-ctx.Done():
				wp.logger.Info("Context cancelled, stopping worker", "worker_id", id)
				return
			}
		}
	}
}
