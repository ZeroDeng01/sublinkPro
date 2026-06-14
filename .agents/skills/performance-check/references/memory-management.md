# Memory Management

This guide covers memory optimization patterns for SublinkPro's Go backend.

## Key Principles

1. Release resources (connections, files, goroutines)
2. Pre-allocate large objects when size is known
3. Pass large structs by pointer
4. Use `strings.Builder` for string concatenation
5. Stream large files instead of loading entirely

## String Building

### Problem: Repeated String Concatenation

**Example violation**:
```go
// ❌ BAD: Repeated allocations in loop
var result string
for _, item := range items {
    result += item + "\n" // Creates new string each iteration!
}
```

### Solution: strings.Builder

```go
// ✅ GOOD: strings.Builder
var builder strings.Builder
builder.Grow(len(items) * 50) // Pre-allocate if size is known
for _, item := range items {
    builder.WriteString(item)
    builder.WriteString("\n")
}
result := builder.String()
```

## Streaming Large Files

### Problem: Loading Entire File Into Memory

**Example violation**:
```go
// ❌ BAD: Loading large file entirely into memory
data, err := os.ReadFile("large_file.json") // May be 100MB+
json.Unmarshal(data, &result)
```

### Solution: Streaming Parser

```go
// ✅ GOOD: Streaming parser
file, err := os.Open("large_file.json")
defer file.Close()
decoder := json.NewDecoder(file)
decoder.Decode(&result) // Processes incrementally
```

## Slice Pre-allocation

### Problem: Repeated Slice Growth

```go
// ❌ BAD: Slice grows repeatedly
var results []Result
for _, item := range items {
    results = append(results, process(item)) // May reallocate multiple times
}
```

### Solution: Pre-allocate Capacity

```go
// ✅ GOOD: Pre-allocate when size is known
results := make([]Result, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}

// ✅ EVEN BETTER: Pre-allocate length if index assignment
results := make([]Result, len(items))
for i, item := range items {
    results[i] = process(item)
}
```

## Map Pre-allocation

```go
// Pre-allocate map capacity
cache := make(map[string]int, 1000)
```

## Pointer vs Value

### When to Use Pointers

```go
// ✅ GOOD: Large struct passed by pointer
type LargeConfig struct {
    // ... many fields ...
}

func ProcessConfig(cfg *LargeConfig) { // Pointer avoids copy
    // ...
}

// ✅ GOOD: Small struct can be passed by value
type Point struct {
    X, Y int
}

func Distance(p1, p2 Point) float64 { // Value is fine
    // ...
}
```

## Resource Cleanup

### Always Use defer for Cleanup

```go
// ✅ GOOD: Ensure resources are released
func ProcessFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close() // Always closed, even on error
    
    // ... process file ...
}

func QueryDatabase() error {
    rows, err := db.Query("SELECT * FROM nodes")
    if err != nil {
        return err
    }
    defer rows.Close() // Always closed
    
    // ... process rows ...
}
```

## Memory Leak Detection

### Common Causes

1. **Goroutine leaks** - Goroutines that never exit
2. **Unclosed resources** - Files, connections, HTTP response bodies
3. **Global caches** - Unbounded maps or slices
4. **Event listeners** - Callbacks that are never unregistered

### Prevention

```go
// ✅ GOOD: Bounded cache with eviction
cache := cache.New(
    5*time.Minute,    // Default expiration
    10*time.Minute,   // Cleanup interval
)
cache.SetMaxSize(1000) // Limit size

// ✅ GOOD: Close HTTP response bodies
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
io.Copy(io.Discard, resp.Body) // Drain body before closing
```

## Checklist

When reviewing memory-sensitive code:
- [ ] **Memory leaks**: Are resources (connections, files, goroutines) released?
- [ ] **Large allocations**: Are large objects (slices, maps) pre-allocated when size is known?
- [ ] **Pointer vs value**: Are large structs passed by pointer?
- [ ] **String concatenation**: Is `strings.Builder` used for building large strings?
- [ ] **JSON marshaling**: Are large objects streamed instead of loaded entirely?
- [ ] **defer statements**: Are cleanup operations deferred?
- [ ] **Bounded caches**: Do caches have size limits?

## Profiling

### Memory Profiling

```bash
# Run with memory profiling
go test -memprofile=mem.prof -bench=.

# Analyze profile
go tool pprof mem.prof
# Commands: top, list, web
```

### Heap Inspection

```go
import "runtime"

var m runtime.MemStats
runtime.ReadMemStats(&m)
log.Printf("Alloc = %v MB", m.Alloc / 1024 / 1024)
log.Printf("TotalAlloc = %v MB", m.TotalAlloc / 1024 / 1024)
log.Printf("Sys = %v MB", m.Sys / 1024 / 1024)
log.Printf("NumGC = %v", m.NumGC)
```

## References

- Go Memory Management: https://go.dev/doc/gc-guide
- Go Performance: https://github.com/dgryski/go-perfbook
