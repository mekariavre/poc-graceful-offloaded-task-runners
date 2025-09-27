# Proof of Concept: Graceful Task Runner

This repo explores using [ants](https://github.com/panjf2000/ants) as a **bounded goroutine pool** for background tasks in Go services.  
The goal is to replace scattered `go func(){}` calls with a **controlled, safe delegator** that:

- Limits concurrency via a worker pool.
- Prevents runaway goroutines & GC pressure.
- Provides a path to **graceful shutdown** so tasks finish or timeout on service exit.

---

## ✨ Why

Currently, async side-effects (logging, metrics, event posting, cleanups) run via raw goroutines.  
This is fast but unsafe — tasks may be lost if the service shuts down abruptly.

Using `ants` gives us:

- Efficient goroutine reuse.
- Bounded task execution.
- A foundation for a centralized background task runner.
