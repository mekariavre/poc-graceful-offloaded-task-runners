# gracego

A lightweight, graceful task delegator for Go, using a worker pool and bounded queue. Ideal for offloading minor tasks with controlled concurrency and graceful shutdown (ensuring dispatched goroutines complete with less risk of data loss).

Sample usecases that will benefit from this: publishing events, logging, and tracking metrics.

## Features

- Bounded concurrency and queue size
- Graceful shutdown with timeout support
- Context-based cancellation
- Simple API for submitting tasks

## Installation

Import the package in your Go project:

```
go get gitgitgit.com/domainscope/gracego
```

## Usage Example

```go
package main

import (
	"context"
	"fmt"
	"time"
	"gitgitgit.com/domainscope/gracego"
)

func main() {
	delegator := gracego.New(4, 10000) // 4 workers, queue size 10000 task items
	delegator.Start()

	for i := 0; i < 20; i++ {
		idx := i
		_ = delegator.Submit(func(ctx context.Context) {
			fmt.Printf("Task %d started\n", idx)
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("Task %d finished\n", idx)
		})
	}

	if err := delegator.Shutdown(); err != nil {
		fmt.Println("Shutdown error:", err)
	}
}
```

## License

MIT
