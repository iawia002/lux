package utils

import (
	"math"
	"sync"
)

type WaitGroupPool struct {
	pool chan int
	wg   *sync.WaitGroup
}

func NewWaitGroupPool(size int) *WaitGroupPool {
	if size <= 0 {
		size = math.MaxInt32
	}
	return &WaitGroupPool{
		pool: make(chan int, size),
		wg:   &sync.WaitGroup{},
	}
}

func (p *WaitGroupPool) Add() {
	p.pool <- 1
	p.wg.Add(1)
}

func (p *WaitGroupPool) Done() {
	<-p.pool
	p.wg.Done()
}

func (p *WaitGroupPool) Wait() {
	p.wg.Wait()
}
