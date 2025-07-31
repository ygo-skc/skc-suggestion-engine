package model

import (
	"sync"
	"sync/atomic"
)

type AtomicWaitGroup[T any] struct {
	data *atomic.Pointer[T]
	wg   *sync.WaitGroup
}

func (t AtomicWaitGroup[T]) Store(d *T) {
	t.data.Store(d)
	t.wg.Done()
}

func NewAtomicWaitGroup[T any](wg *sync.WaitGroup) *AtomicWaitGroup[T] {
	wg.Add(1)
	return &AtomicWaitGroup[T]{data: &atomic.Pointer[T]{}, wg: wg}
}

func (a *AtomicWaitGroup[T]) Load() *T {
	a.wg.Wait()
	return a.data.Load()
}
