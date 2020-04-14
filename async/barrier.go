// Author: lipixun
// Created Time : 2020-04-14 20:12:27
//
// File Name: barrier.go
// Description:
//

package async

import (
	"context"
	"sync"
)

// ConcurrentBarrier implements a concurrent limitation function
type ConcurrentBarrier struct {
	m      sync.Mutex
	max    int
	no     int
	closed bool
	await  AwaitToken
	ch     chan struct{}
}

// NewConcurrentBarrier creates a new ConcurrentBarrier
func NewConcurrentBarrier(max int) *ConcurrentBarrier {
	if max <= 0 {
		panic("Max must be larger than 0")
	}
	ch := make(chan struct{}, max)

	for i := 0; i < max; i++ {
		ch <- struct{}{}
	}
	return &ConcurrentBarrier{
		max: max,
		ch:  ch,
	}
}

// Next tells if next concurrent job should be ran or not
func (b *ConcurrentBarrier) Next(ctx context.Context, opts ...Option) (context.Context, Token, bool) {
	if b.closed {
		return nil, nil, false
	}
	// Wait for one complete
	select {
	case _, ok := <-b.ch:
		if !ok {
			// Closed
			return nil, nil, false
		}
		return b.newRunning(ctx, opts...)
	case <-ctx.Done():
		// Cancel
		return nil, nil, false
	}
}

// Close and wait
func (b *ConcurrentBarrier) Close() {
	if b.closed {
		return
	}
	b.m.Lock()
	if b.closed {
		b.m.Unlock()
		return
	}
	b.closed = true
	b.m.Unlock()
	// Wait for all running job done
	b.await.Wait()
	// Eat all and close the channel
	for {
		select {
		case <-b.ch:
		default:
			close(b.ch)
			return
		}
	}
}

func (b *ConcurrentBarrier) newRunning(ctx context.Context, opts ...Option) (context.Context, Token, bool) {
	if b.closed {
		return nil, nil, false
	}
	b.m.Lock()
	defer b.m.Unlock()
	if b.closed {
		return nil, nil, false
	}
	opts = append(opts, Await(&b.await), cno(b.no), doneCallback(func(*token) {
		b.ch <- struct{}{}
	}))
	b.no++
	actx, token := WithAsync(ctx, opts...)
	return actx, token, true
}
