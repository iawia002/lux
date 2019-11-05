package utils

import (
	"math"
	"sync"
)

// WaitGroupPool pool of WaitGroup
type WaitGroupPool struct {
	pool chan struct{}
	wg   *sync.WaitGroup
}

// NewWaitGroupPool creates a sized pool for WaitGroup
func NewWaitGroupPool(size int) *WaitGroupPool {
	if size <= 0 {
		size = math.MaxInt32
	}
	return &WaitGroupPool{
		pool: make(chan struct{}, size),
		wg:   &sync.WaitGroup{},
	}
}

// Add increments the WaitGroup counter by one.
// See sync.WaitGroup documentation for more information.
func (p *WaitGroupPool) Add() {
	p.pool <- struct{}{}
	p.wg.Add(1)
}

// Done decrements the WaitGroup counter by one.
// See sync.WaitGroup documentation for more information.
func (p *WaitGroupPool) Done() {
	<-p.pool
	p.wg.Done()
}

// Wait blocks until the WaitGroup counter is zero.
// See sync.WaitGroup documentation for more information.
func (p *WaitGroupPool) Wait() {
	p.wg.Wait()
}
