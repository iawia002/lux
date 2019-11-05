package utils

import (
	"sync/atomic"
	"testing"
)

func TestWaitGroupPool(t *testing.T) {
	wgp := NewWaitGroupPool(10)

	var total uint32

	for i := 0; i < 100; i++ {
		wgp.Add()
		go func(total *uint32) {
			defer wgp.Done()
			atomic.AddUint32(total, 1)
		}(&total)
	}
	wgp.Wait()

	if total != 100 {
		t.Fatalf("The size '%d' of the pool did not meet expectations.", total)
	}
}
