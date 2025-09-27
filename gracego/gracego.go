package gracego

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/panjf2000/ants/v2"
)

var ErrQueueFull = errors.New("delegator queue full")

type TaskFunc func(ctx context.Context)

type GracegoDelegator struct {
	pool *ants.Pool

	// ctx    context.Context
	cancel context.CancelFunc
	wgexec sync.WaitGroup

	mxactive sync.Mutex
	active   bool

	onceshutdown  sync.Once
	chbuftasks    chan TaskFunc
	chsigshutdown chan struct{}
	// chflagdrained chan struct{}
}

// New creates a delegator with bounded concurrency and queue size.
func New(concurrency, maxQueueSize int) *GracegoDelegator {
	// ctx, cancel := context.WithCancel(context.Background())

	pool, err := ants.NewPool(
		concurrency,
		ants.WithPreAlloc(true),
	)
	if err != nil {
		err = fmt.Errorf("failed to create worker pool: %w", err)
		panic(err)
	}

	return &GracegoDelegator{
		// ctx:        ctx,
		pool:       pool,
		chbuftasks: make(chan TaskFunc, maxQueueSize),
		// cancel:     cancel,
	}
}

// Start begins dispatchment of tasks into the worker pool.
func (d *GracegoDelegator) Start() {
	// prevent multiple starts
	d.mxactive.Lock()
	if d.active {
		d.mxactive.Unlock()
		return
	}
	d.active = true
	d.mxactive.Unlock()

	// start event loop
	go func() {
		submitTask := func(task TaskFunc) {
			d.wgexec.Add(1)
			_ = d.pool.Submit(func() {
				defer d.wgexec.Done()
				// task(d.ctx)
				task(context.Background())
				log.Println("submit: task executed")
			})
			log.Println("submit: task submitted to pool")
		}
		for {
			select {
			case task := <-d.chbuftasks:
				submitTask(task)
			case <-d.chsigshutdown:
				d.wgexec.Wait()

				d.mxactive.Lock()
				d.active = false
				d.mxactive.Unlock()
				return
			}
		}
	}()

	go func() {
		submitTask := func(task TaskFunc) {
			d.wgexec.Add(1)
			_ = d.pool.Submit(func() {
				defer d.wgexec.Done()
				// task(d.ctx)
				task(context.Background())
				log.Println("submit: task executed")
			})
			log.Println("submit: task submitted to pool")
		}
		for {
			select {
			// case <-d.ctx.Done():
			// 	// drain remaining tasks
			// 	log.Printf("event_loop: will drain %d remaining tasks...\n", len(d.chbuftasks))
			// 	for task := range d.chbuftasks {
			// 		submitTask(task)
			// 	}
			// 	log.Println("event_loop: all remaining tasks drained")
			// 	return
			case task, ok := <-d.chbuftasks:
				if !ok {
					log.Println("event_loop: channel closed")
					return
				}
				log.Println("event_loop: received a task")
				submitTask(task)
			}
		}
	}()
}

// Submit enqueues a task non-blockingly.
// It will return ErrQueueFull if the buffer maximum queue size is full.
func (d *GracegoDelegator) Submit(task TaskFunc) error {
	select {
	// case <-d.ctx.Done():
	// 	return context.Canceled
	case d.chbuftasks <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

// Shutdown gracefully stops accepting tasks and waits for completion or timeout.
// If timeout is zero, it waits until all tasks are completed.
func (d *GracegoDelegator) Shutdown() error {
	var err error
	d.onceshutdown.Do(func() {
		log.Println("shutdown: shutting down delegator...")

		// close channel and end context
		close(d.chbuftasks)
		// d.cancel()

		// wait for all tasks to complete
		done := make(chan struct{})
		go func() {
			log.Println("shutdown: waiting for all tasks to complete...")
			d.wgexec.Wait()
			log.Println("shutdown: all tasks drained")
			d.pool.Release()
			close(done)
		}()

		<-done
	})
	return err
}
