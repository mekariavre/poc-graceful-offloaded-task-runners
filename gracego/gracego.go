package gracego

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

var ErrQueueFull = errors.New("delegator queue full")

type TaskFunc func(ctx context.Context)

type GracegoDelegator struct {
	pool *ants.Pool

	ctx    context.Context
	cancel context.CancelFunc

	wgexec sync.WaitGroup

	mxactive   sync.Mutex
	flagactive bool

	chbuftasks chan TaskFunc
}

// New creates a delegator with bounded concurrency and queue size.
func New(concurrency, maxQueueSize int) *GracegoDelegator {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	ctx, cancel := context.WithCancel(context.Background())

	pool, err := ants.NewPool(
		concurrency,
		ants.WithPreAlloc(true),
	)
	if err != nil {
		err = fmt.Errorf("failed to create worker pool: %w", err)
		panic(err)
	}

	return &GracegoDelegator{
		ctx:    ctx,
		cancel: cancel,

		pool:       pool,
		chbuftasks: make(chan TaskFunc, maxQueueSize),
	}
}

func (d *GracegoDelegator) isactive() bool {
	d.mxactive.Lock()
	defer d.mxactive.Unlock()
	return d.flagactive
}

func (d *GracegoDelegator) setactive(active bool) {
	d.mxactive.Lock()
	defer d.mxactive.Unlock()
	d.flagactive = active
}

// Start begins dispatchment of tasks into the worker pool.
func (d *GracegoDelegator) Start() {
	// prevent multiple starts
	if d.isactive() {
		return
	}
	defer d.setactive(true)

	// start event loop
	go func() {
		delegate := func(task TaskFunc) {
			// log.Println("delegate: delegating task to pool worker")
			_ = d.pool.Submit(func() {
				defer d.wgexec.Done()
				// log.Println("delegate: task execution: started")
				task(d.ctx)
				// log.Println("delegate: task execution: completed")
			})
			// log.Println("delegate: task delegated!")
		}

		for {
			select {
			case task := <-d.chbuftasks:
				// log.Println("event loop: picked new task")
				delegate(task)
			case <-d.ctx.Done():
				// log.Println("event loop: received shutdown signal")
				// log.Printf("event loop: delegating remaining %d tasks", len(d.chbuftasks))
				close(d.chbuftasks)
				for v := range d.chbuftasks {
					// log.Println("event loop: force pushing new task (drain)")
					delegate(v)
				}
				d.wgexec.Wait()
				// log.Println("event loop: all tasks completed")
				return
			}
		}
	}()
}

// Submit enqueues a task non-blockingly.
// It will return ErrQueueFull if the buffer maximum queue size is full.
func (d *GracegoDelegator) Submit(task TaskFunc) error {
	// check if already shutdown
	select {
	case <-d.ctx.Done():
		return context.Canceled
	default:
	}

	// try to enqueue task
	select {
	case d.chbuftasks <- task:
		d.wgexec.Add(1)
		// log.Printf("submit: received new task\n")
		return nil
	default:
		return ErrQueueFull
	}
}

// Shutdown gracefully stops accepting tasks and waits for completion.
func (d *GracegoDelegator) Shutdown() error {
	return d.ShutdownWithTimeout(-1)
}

// Shutdown gracefully stops accepting tasks and waits for completion or timeout.
// If timeout is zero, it cancels all running tasks immediately.
// If timeout is negative, it waits until all tasks are completed.
func (d *GracegoDelegator) ShutdownWithTimeout(timeout time.Duration) error {
	ctx := context.Background()
	switch {
	case timeout > 0:
		// wait with timeout
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	case timeout == 0:
		// immediate cancel
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(context.Background())
		cancel()
	}

	defer d.pool.Release() // ensure free pool resources

	// channelize a wgexec wait
	done := make(chan struct{})
	go func() {
		d.cancel() // cancel internal context
		d.wgexec.Wait()
		close(done)
		// log.Printf("shutdown: all tasks completed\n")
	}()

	// wait for signals
	select {
	case <-done:
		// log.Printf("shutdown: done gracefully\n")
		return nil
	case <-ctx.Done():
		// log.Printf("shutdown: timeout reached, abrupt shutdown\n")
		return ctx.Err()
	}
}
