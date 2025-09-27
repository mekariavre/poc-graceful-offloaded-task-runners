package gracego

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// should create a valid delegator
	t.Run("ok", func(t *testing.T) {
		out := New(2, 10)
		assert.NotNil(t, out)
	})

	// should panic on invalid params
	t.Run("panic on invalid params", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = New(-10, 10) // invalid concurrency
		})
	})
}

func TestGracegoDelegator_Executions(t *testing.T) {
	// should start fine
	t.Run("ok", func(t *testing.T) {
		out := New(2, 10)
		require.NotNil(t, out)

		out.Start()
	})

	// should execute tasks and wait for completion on shutdown
	t.Run("execute tasks", func(t *testing.T) {
		t0 := time.Now()
		log.Printf("start time: %s\n", t0.Format(time.RFC3339Nano))
		out := New(2, 250)
		require.NotNil(t, out)
		out.Start()

		ctr := &counter{}
		for i := 0; i < 250; i++ {
			err := out.Submit(func(ctx context.Context) {
				time.Sleep(1 * time.Millisecond) // simulate work
				ctr.inc()                        // simulate work
			})
			assert.NoError(t, err)
		}

		// should not accept new tasks after shutdown
		err := out.Shutdown()
		assert.NoError(t, err)

		// should have all tasks executed
		assert.Equal(t, 250, ctr.count())
	})

	// should return error when queue is full
	t.Run("queue full", func(t *testing.T) {
		out := New(1, 3) // small queue
		require.NotNil(t, out)
		out.Start()

		submitTask := func() error {
			return out.Submit(func(ctx context.Context) {
				time.Sleep(5 * time.Millisecond) // simulate work
			})
		}

		errcount := 0
		inconerr := func(err error) {
			if err != nil && err == ErrQueueFull {
				errcount++
			}
		}

		inconerr(submitTask()) // 1
		inconerr(submitTask()) // 2
		inconerr(submitTask()) // 3
		inconerr(submitTask()) // 4
		inconerr(submitTask()) // 5

		require.NotZero(t, errcount)
	})

	// should return error when submitting after shutdown
	t.Run("submit after shutdown", func(t *testing.T) {
		del := New(2, 5)
		require.NotNil(t, del)
		del.Start()
		require.NoError(t, del.Shutdown())

		err := del.Submit(func(ctx context.Context) {})
		assert.ErrorIs(t, err, context.Canceled)
	})

	// coverage boost: double start
	t.Run("double start", func(t *testing.T) {
		out := New(2, 5)
		require.NotNil(t, out)
		out.Start()
		out.Start() // should be no-op
		require.NoError(t, out.Shutdown())
	})

	// coverage boost: task drained before shutdown
	t.Run("drain before shutdown", func(t *testing.T) {
		out := New(2, 5) // small queue
		require.NotNil(t, out)
		out.Start()

		submitTask := func(dur time.Duration) error {
			return out.Submit(func(ctx context.Context) {
				time.Sleep(dur) // simulate work
			})
		}

		assert.NoError(t, submitTask(1*time.Millisecond)) // 1
		assert.NoError(t, submitTask(1*time.Millisecond)) // 2

		time.Sleep(2 * time.Millisecond) // wait for tasks to be drained

		require.NoError(t, out.Shutdown())
	})
}

// helper

type counter struct {
	sm sync.Map
}

func (c *counter) inc() {
	randkey := gonanoid.Must(10)
	c.sm.Store(randkey, struct{}{})
	// log.Printf("counter inc: %s total: %d\n", randkey, c.count())
}

func (c *counter) count() int {
	count := 0
	c.sm.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

func TestGracegoDelegator_ShutdownWithTimeout(t *testing.T) {
	makedel := func() *GracegoDelegator {
		del := New(2, 50)
		require.NotNil(t, del)
		return del
	}

	// should shutdown fine with timeout
	t.Run("shutdown with timeout", func(t *testing.T) {
		del := makedel()

		for i := 0; i < 50; i++ {
			err := del.Submit(func(ctx context.Context) {
				time.Sleep(1 * time.Second) // simulate work
			})
			assert.NoError(t, err)
		}

		// should not accept new tasks after shutdown
		del.Start()
		err := del.ShutdownWithTimeout(1 * time.Millisecond)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	// should shutdown fine with zero timeout (cancel immediately)
	t.Run("shutdown with zero timeout should cancel immediately", func(t *testing.T) {
		del := makedel()

		ctr := &counter{}
		for i := 0; i < 5; i++ {
			err := del.Submit(func(ctx context.Context) {
				time.Sleep(1 * time.Second) // simulate work
				ctr.inc()                   // simulate work
			})
			assert.NoError(t, err)
		}

		// should not accept new tasks after shutdown
		del.Start()
		err := del.ShutdownWithTimeout(0)
		assert.ErrorIs(t, err, context.Canceled)

		// should have some tasks executed (not all, since we cancelled immediately)
		assert.Equal(t, ctr.count(), 0)
	})

	// should shutdown fine with negative timeout (wait indefinitely)
	t.Run("shutdown with negative timeout", func(t *testing.T) {
		del := makedel()

		ctr := &counter{}
		for i := 0; i < 5; i++ {
			err := del.Submit(func(ctx context.Context) {
				time.Sleep(1 * time.Millisecond) // simulate work
				ctr.inc()                        // simulate work
			})
			assert.NoError(t, err)
		}

		// should not accept new tasks after shutdown
		del.Start()
		err := del.ShutdownWithTimeout(-1)
		assert.NoError(t, err)

		// should have all tasks executed
		assert.Equal(t, 5, ctr.count())
	})
}
