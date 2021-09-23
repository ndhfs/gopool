package gopool

import (
	"context"
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/jitter"
	"github.com/Rican7/retry/strategy"
	"math/rand"
	"sync"
	"time"
)

type WorkersPool struct {
	mu sync.RWMutex
	size int
	kill chan struct{}
	wg sync.WaitGroup
	worker Worker
}

type Worker func(attempt uint, ctx context.Context) error

func NewWorkersPool() *WorkersPool {
	return &WorkersPool{kill: make(chan struct{})}
}

func (p *WorkersPool) Do(size int, worker Worker) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.worker = worker
	for p.size < size {
		p.runWorker()
		p.size++
	}
}

func (p *WorkersPool) Resize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if size == p.size {
		return
	}

	for p.size < size {
		p.size++
		p.runWorker()
	}

	for p.size > size {
		p.size--
		p.kill <- struct{}{}
	}
}

func (p *WorkersPool) runWorker() {
	p.wg.Add(1)
	go func() {
		ctx, doneFn := context.WithCancel(context.Background())
		defer func() {
			p.wg.Done()
		}()
		go func() {
			select {
			case <-p.kill:
				doneFn()
			}
		}()

		retry.Retry(
			func(attempt uint) error {
				return p.worker(attempt, ctx)
			},
			func(attempt uint) bool {
				select {
				case <-ctx.Done():
					return false
				default:
					return true
				}
			},
			strategy.BackoffWithJitter(
				backoff.BinaryExponential(10*time.Millisecond),
				jitter.Deviation(rand.New(rand.NewSource(time.Now().UnixNano())), 0.5),
			),
		)
	}()
}

func (p *WorkersPool) Stop() {
	p.mu.Lock()
	for p.size != 0 {
		p.size--
		p.kill <- struct{}{}
	}
	p.mu.Unlock()
	p.wg.Wait()
}
