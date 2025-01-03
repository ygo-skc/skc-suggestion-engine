package util

import (
	"context"
	"sync"
)

type Task interface {
	Process()
}

type WorkerPool struct {
	tasks   []Task
	workers int
	ctx     context.Context
	tChan   chan Task
}

type WPOption func(*WorkerPool)

func WithWorkers(workers int) WPOption {
	return func(s *WorkerPool) {
		s.workers = workers
	}
}

func WithContext(ctx context.Context) WPOption {
	return func(s *WorkerPool) {
		s.ctx = ctx
	}
}

func NewWorkerPool(tasks []Task, options ...WPOption) *WorkerPool {
	// default values
	wp := &WorkerPool{
		tasks:   tasks,
		workers: 5,
		ctx:     context.TODO(),
		tChan:   make(chan Task, len(tasks)),
	}

	// override with custom values from caller
	for _, option := range options {
		option(wp)
	}

	return wp
}

func (wp *WorkerPool) worker(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case task, ok := <-wp.tChan:
			if !ok {
				return
			}
			task.Process()
		}
	}
}

func (wp *WorkerPool) Run() {
	wg := sync.WaitGroup{}

	for i := 0; i < wp.workers; i++ {
		wg.Add(1)
		go wp.worker(&wg)
	}

	for _, task := range wp.tasks {
		wp.tChan <- task
	}

	close(wp.tChan)
	wg.Wait()
}
