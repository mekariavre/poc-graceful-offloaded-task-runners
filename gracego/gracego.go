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
			log.Println("delegate: delegating task to pool worker")
			_ = d.pool.Submit(func() {
				defer d.wgexec.Done()
				log.Println("delegate: task execution: started")
				task(d.ctx)
				log.Println("delegate: task execution: completed")
			})
			log.Println("delegate: task delegated!")
		}

		for {
			select {
			case task := <-d.chbuftasks:
				log.Println("event loop: picked new task")
				delegate(task)
			case <-d.ctx.Done():
				log.Println("event loop: received shutdown signal")
				log.Printf("event loop: delegating remaining %d tasks", len(d.chbuftasks))
				close(d.chbuftasks)
				for v := range d.chbuftasks {
					log.Println("event loop: force pushing new task (drain)")
					delegate(v)
				}
				d.wgexec.Wait()
				d.pool.Release()
				log.Println("event loop: all tasks completed")
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
		log.Printf("submit: received new task\n")
		return nil
	default:
		return ErrQueueFull
	}
}

// Shutdown gracefully stops accepting tasks and waits for completion or timeout.
// If timeout is zero, it waits until all tasks are completed.
func (d *GracegoDelegator) Shutdown() error {
	log.Printf("shutdown: fired\n")
	d.cancel()
	log.Printf("shutdown: waiting for exec done\n")
	d.wgexec.Wait()
	log.Printf("shutdown: exec done\n")
	return nil
}
