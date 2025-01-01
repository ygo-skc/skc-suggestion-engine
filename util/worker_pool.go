package util

import (
	"context"
	"sync"
)

type Task interface {
	Process()
}

type WorkerPool struct {
	Tasks   []Task
	Workers int
	Context context.Context
	tChan   chan Task
}

func (wp *WorkerPool) worker(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-wp.Context.Done():
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
	wp.tChan = make(chan Task, len(wp.Tasks))

	for i := 0; i < wp.Workers; i++ {
		wg.Add(1)
		go wp.worker(&wg)
	}

	for _, task := range wp.Tasks {
		wp.tChan <- task
	}

	close(wp.tChan)
	wg.Wait()
}
