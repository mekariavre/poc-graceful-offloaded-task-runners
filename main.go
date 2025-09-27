package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

func main() {
	t0 := time.Now()

	fmt.Println("starting...")
	scenario_5()
	fmt.Printf("exited (%s)\n", time.Since(t0))
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
			fmt.Printf("Task %d running\n", idx)
			time.Sleep(time.Second)
		})
	}

	// wait so we can see output before exit
	time.Sleep(3 * time.Second)
}
