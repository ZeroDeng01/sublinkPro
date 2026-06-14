# Concurrency & Parallelization

This guide covers Go concurrency patterns for SublinkPro.

## Key Principles

1. Make blocking operations asynchronous
2. Run independent tasks concurrently
3. Prevent goroutine leaks
4. Protect shared resources with mutexes
5. Use buffered channels appropriately
6. Limit concurrency to avoid resource exhaustion

## Parallel Processing

### Problem: Sequential Processing

**Example violation**:
```go
// ❌ BAD: Sequential processing
func TestMultipleNodes(nodeIDs []uint) []TestResult {
    results := []TestResult{}
    for _, id := range nodeIDs {
        result := testNode(id) // Blocks for each node (could take 5s each!)
        results = append(results, result)
    }
    return results
}
```

### Solution: Goroutines with Semaphore

```go
// ✅ GOOD: Parallel processing with semaphore to limit concurrency

package examples

import "sync"

type TestResult struct {
	NodeID uint
	Delay  int
	Error  error
}

// Parallel node testing with concurrency limit
func TestMultipleNodes(nodeIDs []uint) []TestResult {
	results := make([]TestResult, len(nodeIDs))
	var wg sync.WaitGroup

	// Limit concurrency to avoid overwhelming resources
	semaphore := make(chan struct{}, 10) // Max 10 concurrent

	for i, id := range nodeIDs {
		wg.Add(1)
		go func(index int, nodeID uint) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			results[index] = testNode(nodeID)
		}(i, id)
	}

	wg.Wait()
	return results
}

func testNode(nodeID uint) TestResult {
	// Simulate network delay test
	return TestResult{NodeID: nodeID, Delay: 100}
}
```

## Preventing Goroutine Leaks

### Problem: Goroutine Leak

**Example violation**:
```go
// ❌ BAD: Goroutine leak
func StreamData(w http.ResponseWriter) {
    ch := make(chan Data)
    go producer(ch) // If consumer exits early, producer leaks!
    
    for data := range ch {
        if someCondition {
            return // Goroutine never stops!
        }
        w.Write(data)
    }
}
```

### Solution: Context-Based Cleanup

```go
// ✅ GOOD: Proper goroutine cleanup with context
func StreamData(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithCancel(r.Context())
    defer cancel() // Ensure cleanup
    
    ch := make(chan Data)
    go producer(ctx, ch) // Producer respects context
    
    for {
        select {
        case data := <-ch:
            if someCondition {
                return // cancel() will stop producer
            }
            w.Write(data)
        case <-ctx.Done():
            return
        }
    }
}
```

## Race Condition Prevention

### Problem: Unsynchronized Access

```go
// ❌ BAD: Race condition
var counter int
for i := 0; i < 1000; i++ {
    go func() {
        counter++ // Race!
    }()
}
```

### Solution: Mutex or Atomic Operations

```go
// ✅ GOOD: Synchronized with mutex
var (
    counter int
    mu      sync.Mutex
)

for i := 0; i < 1000; i++ {
    go func() {
        mu.Lock()
        counter++
        mu.Unlock()
    }()
}

// ✅ BETTER: Atomic operations for simple counters
var counter int64
for i := 0; i < 1000; i++ {
    go func() {
        atomic.AddInt64(&counter, 1)
    }()
}
```

## Channel Patterns

### Buffered vs Unbuffered

```go
// Unbuffered: Sender blocks until receiver reads
ch := make(chan int)

// Buffered: Sender blocks only when buffer is full
ch := make(chan int, 100)
```

### Fan-Out, Fan-In

```go
// Fan-out: Multiple workers processing from one channel
func fanOut(jobs <-chan Job, workers int) []<-chan Result {
    results := make([]<-chan Result, workers)
    for i := 0; i < workers; i++ {
        results[i] = worker(jobs)
    }
    return results
}

// Fan-in: Merge multiple channels into one
func fanIn(channels ...<-chan Result) <-chan Result {
    out := make(chan Result)
    var wg sync.WaitGroup
    
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan Result) {
            defer wg.Done()
            for result := range c {
                out <- result
            }
        }(ch)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```

## Checklist

When reviewing concurrent code:
- [ ] **Blocking operations**: Are long operations executed asynchronously?
- [ ] **Parallel processing**: Can independent tasks run concurrently?
- [ ] **Goroutine leaks**: Are goroutines properly cleaned up (use context)?
- [ ] **Race conditions**: Are shared resources protected with mutexes or atomic operations?
- [ ] **Channel buffering**: Are channels sized appropriately?
- [ ] **Concurrency limits**: Is there a semaphore or worker pool to prevent resource exhaustion?

## Common Patterns in SublinkPro

1. **Speed testing** - Parallel node testing with semaphore (see `services/scheduler/speedtest_task.go`)
2. **Subscription fetching** - Concurrent fetches with timeout
3. **DNS resolution** - Parallel resolution for multiple domains
4. **Batch operations** - Worker pool pattern for bulk updates

## Profiling

Detect race conditions:
```bash
go test -race ./...
go build -race
```

Detect goroutine leaks:
```go
import "runtime"

before := runtime.NumGoroutine()
// ... run test ...
after := runtime.NumGoroutine()
if after > before {
    t.Errorf("Goroutine leak: %d -> %d", before, after)
}
```

## References

- `services/scheduler/speedtest_task.go` - Concurrent node testing example
- Go Concurrency Patterns: https://go.dev/blog/pipelines
