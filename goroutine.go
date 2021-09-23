package gopool

import (
	"context"
	"time"
)

type GoPool struct {
	sem chan struct{}
}

func NewGoPool(limit int) *GoPool {
	return &GoPool{sem: make(chan struct{}, limit)}
}

func (g *GoPool) Do(fn func()) error {
	select {
	case g.sem <- struct{}{}:
	default:
		return context.DeadlineExceeded
	}
	return g.do(fn)
}

func (g *GoPool) DoWithTimeout(waitTimeout time.Duration, f func()) error {
	ctx, closeFn := context.WithTimeout(context.Background(), waitTimeout)
	defer closeFn()
	return g.DoWithContext(ctx, f)
}

func (g *GoPool) DoWithContext(ctx context.Context, fn func()) error {
	select {
	case g.sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	return g.do(fn)
}

func (g *GoPool) do(fn func()) error {
	go func() {
		defer func() {
			<-g.sem
		}()
		fn()
	}()
	return nil
}
