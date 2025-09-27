package main

import (
	"fmt"
	"time"

	"github.com/panjf2000/ants/v2"
)

func main() {
	t0 := time.Now()

	fmt.Println("starting...")
	scenario_1()
	fmt.Printf("exited (%s)\n", time.Since(t0))
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
