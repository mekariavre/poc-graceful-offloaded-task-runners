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

// func TestNew(t *testing.T) {
// 	// should create a valid delegator
// 	t.Run("ok", func(t *testing.T) {
// 		out := New(2, 10)
// 		assert.NotNil(t, out)
// 	})

// 	// should panic on invalid params
// 	t.Run("panic on invalid params", func(t *testing.T) {
// 		assert.Panics(t, func() {
// 			_ = New(-10, 10) // invalid concurrency
// 		})
// 	})
// }

func TestGracegoDelegator_Executions(t *testing.T) {
	// // should start fine
	// t.Run("ok", func(t *testing.T) {
	// 	out := New(2, 10)
	// 	require.NotNil(t, out)

	// 	out.Start()
	// })

	// should execute tasks and wait for completion on shutdown
	t.Run("execute tasks", func(t *testing.T) {
		out := New(5, 25)
		require.NotNil(t, out)
		out.Start()

		ctr := &counter{}
		for i := 0; i < 25; i++ {
			log.Printf("publishing task %d\n", i)
			err := out.Submit(func(ctx context.Context) {
				counter := i
				time.Sleep(1 * time.Millisecond) // simulate work
				ctr.inc()                        // simulate work
				log.Printf("has run task %d\n", counter)
			})
			assert.NoError(t, err)
		}

		// should not accept new tasks after shutdown
		err := out.Shutdown()
		assert.NoError(t, err)

		// should have all tasks executed
		log.Println("all tasks processed, checking results...")
		assert.Equal(t, 25, ctr.count())
	})

	// // should return error when queue is full
	// t.Run("queue full", func(t *testing.T) {
	// 	out := New(2, 5) // small queue
	// 	require.NotNil(t, out)
	// 	out.Start()

	// 	submitTask := func() error {
	// 		return out.Submit(func(ctx context.Context) {
	// 			time.Sleep(1 * time.Millisecond) // simulate work
	// 		})
	// 	}

	// 	assert.NoError(t, submitTask())               // 1
	// 	assert.NoError(t, submitTask())               // 2
	// 	assert.NoError(t, submitTask())               // 3
	// 	assert.NoError(t, submitTask())               // 4
	// 	assert.NoError(t, submitTask())               // 5
	// 	assert.ErrorIs(t, submitTask(), ErrQueueFull) // 6 - should be full now

	// 	require.NoError(t, out.Shutdown())
	// })

	// // coverage boost: task drained before shutdown
	// t.Run("drain before shutdown", func(t *testing.T) {
	// 	out := New(2, 5) // small queue
	// 	require.NotNil(t, out)
	// 	out.Start()

	// 	submitTask := func(dur time.Duration) error {
	// 		return out.Submit(func(ctx context.Context) {
	// 			time.Sleep(dur) // simulate work
	// 		})
	// 	}

	// 	assert.NoError(t, submitTask(1*time.Millisecond)) // 1
	// 	assert.NoError(t, submitTask(1*time.Millisecond)) // 2

	// 	time.Sleep(2 * time.Millisecond) // wait for tasks to be drained

	// 	require.NoError(t, out.Shutdown())
	// })
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
