package main

import (
	"fmt"
	"log"
	"time"

	"github.com/panjf2000/ants/v2"
)

func main() {
	t0 := time.Now()

	fmt.Println("starting...")
	scenario_3()
	fmt.Printf("exited (%s)\n", time.Since(t0))
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
