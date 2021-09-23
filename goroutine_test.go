package gopool

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGoPool_Do(t *testing.T) {
	gp := NewGoPool(1)
	gp.Do(func() {
		time.Sleep(1 * time.Second)
	})
	err := gp.Do(func() {
		time.Sleep(1 * time.Second)
	})
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestGoPool_DoWithTimeout(t *testing.T) {
	{
		gp := NewGoPool(1)
		gp.Do(func() {
			time.Sleep(2 * time.Second)
		})
		err := gp.DoWithTimeout(1*time.Second, func() {
			time.Sleep(1 * time.Second)
		})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	}
	{
		gp := NewGoPool(1)
		gp.Do(func() {
			time.Sleep(1 * time.Second)
		})
		err := gp.DoWithTimeout(1*time.Second, func() {
			time.Sleep(1 * time.Second)
		})
		assert.NoError(t, err)
	}
}

func TestGoPool_DoWithContext(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	gp := NewGoPool(1)
	gp.Do(func() {
		time.Sleep(2 * time.Second)
	})

	cancelFn()
	err := gp.DoWithContext(ctx, func() {
		time.Sleep(1 * time.Second)
	})
	assert.ErrorIs(t, err, context.Canceled)
}
