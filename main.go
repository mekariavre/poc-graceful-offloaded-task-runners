package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mekariavre/poc-graceful-offloaded-task-runners/gracego"
	"github.com/panjf2000/ants/v2"
)

func main() {
	t0 := time.Now()

	fmt.Println("starting...")
	scenario_6()
	log.Printf("exited (%s)\n", time.Since(t0))
}

func scenario_6() {
	delegator := gracego.New(4, 10000) // 4 workers, queue size 10000 task items
	delegator.Start()

	for i := 0; i < 10; i++ {
		idx := i
		_ = delegator.Submit(func(ctx context.Context) {
			log.Printf("task (light) %d started\n", idx)
			time.Sleep(100 * time.Millisecond)
			log.Printf("task (light) %d finished\n", idx)
		})
	}

	// submit some heavy tasks
	for i := 0; i < 5; i++ {
		idx := i
		_ = delegator.Submit(func(ctx context.Context) {
			log.Printf("task (heavy) %d started\n", idx)
			time.Sleep(2 * time.Second)
			log.Printf("task (heavy) %d finished\n", idx)
		})
	}

	// doing other stuff
	log.Println("doing other stuff here!")

	if err := delegator.Shutdown(); err != nil {
		log.Println("Shutdown error:", err)
	}

	// starting...
	// 2025/09/27 17:05:48.500896 doing other stuff here!
	// 2025/09/27 17:05:48.500995 task (light) 3 started
	// 2025/09/27 17:05:48.501001 task (light) 2 started
	// 2025/09/27 17:05:48.501004 task (light) 0 started
	// 2025/09/27 17:05:48.501031 task (light) 1 started
	// 2025/09/27 17:05:48.602101 task (light) 0 finished
	// 2025/09/27 17:05:48.602125 task (light) 4 started
	// 2025/09/27 17:05:48.602128 task (light) 3 finished
	// 2025/09/27 17:05:48.602129 task (light) 5 started
	// 2025/09/27 17:05:48.602132 task (light) 2 finished
	// 2025/09/27 17:05:48.602131 task (light) 1 finished
	// 2025/09/27 17:05:48.602137 task (light) 7 started
	// 2025/09/27 17:05:48.602133 task (light) 6 started
	// 2025/09/27 17:05:48.703160 task (light) 6 finished
	// 2025/09/27 17:05:48.703178 task (light) 8 started
	// 2025/09/27 17:05:48.703148 task (light) 5 finished
	// 2025/09/27 17:05:48.703185 task (light) 9 started
	// 2025/09/27 17:05:48.703168 task (light) 4 finished
	// 2025/09/27 17:05:48.703192 task (heavy) 0 started
	// 2025/09/27 17:05:48.703170 task (light) 7 finished
	// 2025/09/27 17:05:48.703208 task (heavy) 1 started
	// 2025/09/27 17:05:48.804210 task (light) 8 finished
	// 2025/09/27 17:05:48.804232 task (light) 9 finished
	// 2025/09/27 17:05:48.804242 task (heavy) 2 started
	// 2025/09/27 17:05:48.804272 task (heavy) 3 started
	// 2025/09/27 17:05:50.703943 task (heavy) 0 finished
	// 2025/09/27 17:05:50.703940 task (heavy) 1 finished
	// 2025/09/27 17:05:50.703964 task (heavy) 4 started
	// 2025/09/27 17:05:50.805293 task (heavy) 3 finished
	// 2025/09/27 17:05:50.805314 task (heavy) 2 finished
	// 2025/09/27 17:05:52.705098 task (heavy) 4 finished
	// 2025/09/27 17:05:52.705269 exited (4.204364416s)
}

func scenario_5() {
	fmt.Println("scenario: non-blocking with waiting without worker group")

	// create a pool with 5 workers, non-blocking
	pool, _ := ants.NewPool(5, ants.WithNonblocking(true))
	defer pool.Release()

	taskCount := 10
	done := make(chan struct{}, taskCount)

	for i := 0; i < taskCount; i++ {
		idx := i
		_ = pool.Submit(func() {
			log.Printf("Task %d starting\n", idx)
			time.Sleep(2 * time.Second)
			log.Printf("Task %d done\n", idx)
			done <- struct{}{}
		})
	}

	log.Println("doing other stuff")

	// wait for all tasks to complete
	for i := 0; i < taskCount; i++ {
		<-done
	}
	log.Println("All tasks completed")

	// starting...
	// scenario: non-blocking with waiting without worker group
	// 2025/09/27 09:30:52 doing other stuff
	// 2025/09/27 09:30:52 Task 3 starting
	// 2025/09/27 09:30:52 Task 1 starting
	// 2025/09/27 09:30:52 Task 0 starting
	// 2025/09/27 09:30:52 Task 2 starting
	// 2025/09/27 09:30:52 Task 4 starting
	// 2025/09/27 09:30:54 Task 4 done
	// 2025/09/27 09:30:54 Task 2 done
	// 2025/09/27 09:30:54 Task 0 done
	// 2025/09/27 09:30:54 Task 3 done
	// 2025/09/27 09:30:54 Task 1 done
}

func scenario_4() {
	fmt.Println("scenario: non-blocking with waiting using worker group")

	// create a pool with 5 workers, non-blocking
	pool, _ := ants.NewPool(0, ants.WithNonblocking(true))
	defer pool.Release()

	var wg sync.WaitGroup
	taskCount := 10

	for i := 0; i < taskCount; i++ {
		idx := i
		wg.Add(1)
		// err must be ignored, because with non-blocking mode (https://pkg.go.dev/github.com/panjf2000/ants/v2#Options)
		_ = pool.Submit(func() {
			defer wg.Done()
			log.Printf("Task %d starting\n", idx)
			time.Sleep(2 * time.Second)
			log.Printf("Task %d done\n", idx)
		})
	}

	// doing other stuff
	log.Println("doing other stuff")

	// wait for all tasks to complete
	wg.Wait()
	log.Println("All tasks completed")

	// starting...
	// scenario: non-blocking with waiting
	// 2025/09/27 09:27:49 Task 4 starting
	// 2025/09/27 09:27:49 Task 1 starting
	// 2025/09/27 09:27:49 Task 5 starting
	// 2025/09/27 09:27:49 Task 6 starting
	// 2025/09/27 09:27:49 Task 7 starting
	// 2025/09/27 09:27:49 Task 8 starting
	// 2025/09/27 09:27:49 Task 3 starting
	// 2025/09/27 09:27:49 Task 0 starting
	// 2025/09/27 09:27:49 Task 2 starting
	// 2025/09/27 09:27:49 doing other stuff
	// 2025/09/27 09:27:49 Task 9 starting
	// 2025/09/27 09:27:51 Task 9 done
	// 2025/09/27 09:27:51 Task 5 done
	// 2025/09/27 09:27:51 Task 6 done
	// 2025/09/27 09:27:51 Task 7 done
	// 2025/09/27 09:27:51 Task 8 done
	// 2025/09/27 09:27:51 Task 3 done
	// 2025/09/27 09:27:51 Task 0 done
	// 2025/09/27 09:27:51 Task 2 done
	// 2025/09/27 09:27:51 Task 4 done
	// 2025/09/27 09:27:51 Task 1 done
	// 2025/09/27 09:27:51 All tasks completed
	// exited (2.002275667s)
}

func scenario_4a() {
	fmt.Println("scenario: non-blocking with waiting using worker group (with data over closure)")

	// create a pool with 5 workers, non-blocking
	pool, _ := ants.NewPool(0, ants.WithNonblocking(true))
	defer pool.Release()

	var wg sync.WaitGroup
	taskCount := 10

	counter := 0
	for i := 0; i < taskCount; i++ {
		idx := i
		wg.Add(1)
		inLoopCounter := counter
		// err must be ignored, because with non-blocking mode (https://pkg.go.dev/github.com/panjf2000/ants/v2#Options)
		_ = pool.Submit(func() {
			counter++
			internalCounter := counter

			defer wg.Done()
			log.Printf("Task %d starting\n", idx)
			time.Sleep(2 * time.Second)
			log.Printf("Task %d done (counter: %d; in_loop_counter: %d; internal_counter: %d)\n", idx, counter, inLoopCounter, internalCounter)
		})
	}

	// doing other stuff
	log.Println("doing other stuff")

	// wait for all tasks to complete
	wg.Wait()
	log.Println("All tasks completed")

	// starting...
	// scenario: non-blocking with waiting using worker group (with data over closure)
	// 2025/09/27 09:38:07 Task 0 starting
	// 2025/09/27 09:38:07 Task 7 starting
	// 2025/09/27 09:38:07 Task 9 starting
	// 2025/09/27 09:38:07 Task 4 starting
	// 2025/09/27 09:38:07 Task 5 starting
	// 2025/09/27 09:38:07 Task 6 starting
	// 2025/09/27 09:38:07 Task 1 starting
	// 2025/09/27 09:38:07 Task 8 starting
	// 2025/09/27 09:38:07 Task 2 starting
	// 2025/09/27 09:38:07 Task 3 starting
	// 2025/09/27 09:38:07 doing other stuff
	// 2025/09/27 09:38:09 Task 4 done (counter: 10; in_loop_counter: 0; internal_counter: 3)
	// 2025/09/27 09:38:09 Task 6 done (counter: 10; in_loop_counter: 1; internal_counter: 7)
	// 2025/09/27 09:38:09 Task 3 done (counter: 10; in_loop_counter: 0; internal_counter: 10)
	// 2025/09/27 09:38:09 Task 1 done (counter: 10; in_loop_counter: 0; internal_counter: 8)
	// 2025/09/27 09:38:09 Task 8 done (counter: 10; in_loop_counter: 1; internal_counter: 6)
	// 2025/09/27 09:38:09 Task 2 done (counter: 10; in_loop_counter: 0; internal_counter: 9)
	// 2025/09/27 09:38:09 Task 5 done (counter: 10; in_loop_counter: 1; internal_counter: 4)
	// 2025/09/27 09:38:09 Task 7 done (counter: 10; in_loop_counter: 1; internal_counter: 5)
	// 2025/09/27 09:38:09 Task 9 done (counter: 10; in_loop_counter: 1; internal_counter: 2)
	// 2025/09/27 09:38:09 Task 0 done (counter: 10; in_loop_counter: 0; internal_counter: 1)
	// 2025/09/27 09:38:09 All tasks completed
	// exited (2.001664584s)
}

func scenario_3a() {
	fmt.Println("scenario: without waiting & non-blocking (with data over closure)")
	// create a pool with 5 workers
	pool, _ := ants.NewPool(5, ants.WithNonblocking(true))
	defer pool.Release()

	// submit tasks non-blocking, ignore error if pool is full
	counter := 0
	for i := 0; i < 20; i++ {
		idx := i
		_ = pool.Submit(func() {
			counter++
			time.Sleep(5000 * time.Millisecond)
			log.Printf("Task %d done (counter: %d)\n", idx)
		})
	}

	// immediately return, not waiting for tasks
	log.Println("main function continues immediately")

	// starting...
	// scenario: without waiting & non-blocking (with data over closure)
	// 2025/09/27 09:35:07 main function continues immediately
	// exited (122.75µs)
}

func scenario_3() {
	fmt.Println("scenario: without waiting & non-blocking")
	// create a pool with 5 workers
	pool, _ := ants.NewPool(5, ants.WithNonblocking(true))
	defer pool.Release()

	// submit tasks non-blocking, ignore error if pool is full
	for i := 0; i < 20; i++ {
		idx := i
		_ = pool.Submit(func() {
			time.Sleep(5000 * time.Millisecond)
			log.Printf("Task %d done\n", idx)
		})
	}

	// immediately return, not waiting for tasks
	log.Println("main function continues immediately")

	// starting...
	// scenario: without waiting & non-blocking
	// 2025/09/27 09:11:07 main function continues immediately
	// exited (128.75µs)
}

func scenario_2() {
	fmt.Println("scenario: without waiting")
	// create a pool with 5 workers
	pool, _ := ants.NewPool(5)
	defer pool.Release()

	// submit tasks
	for i := 0; i < 20; i++ {
		idx := i
		_ = pool.Submit(func() {
			time.Sleep(5000 * time.Millisecond)
			log.Printf("Task %d done\n", idx)
		})
	}

	// doing other stuff
	log.Println("doing other stuff")
}

func scenario_1() {
	fmt.Println("scenario: basic usage")
	// create a pool with 10 workers
	pool, _ := ants.NewPool(10)
	defer pool.Release()

	// submit tasks
	for i := 0; i < 20; i++ {
		idx := i
		_ = pool.Submit(func() {
			log.Printf("Task %d running\n", idx)
			time.Sleep(time.Second)
		})
	}

	// wait so we can see output before exit
	time.Sleep(3 * time.Second)
}
