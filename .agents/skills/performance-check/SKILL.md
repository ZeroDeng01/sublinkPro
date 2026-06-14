---
name: performance-check
description: "Performance review checklist for optimization, scalability, and resource usage. Use when changes affect queries, APIs, rendering, caching, or algorithms. Not for simple config changes."
version: "1.0.0"
author: "SublinkPro Team"
user-invocable: true
---

# Performance Check Skill

This skill provides a comprehensive performance review checklist for code changes that may impact application performance, scalability, or resource usage.

**When to use**: When making changes to:
- Database queries and models
- API endpoints (especially list/search operations)
- Frontend rendering logic
- Background jobs and scheduled tasks
- Caching mechanisms
- Large data processing
- File operations
- Network requests
- Recursive algorithms

---

## Performance Check Checklist

### 🗄️ 1. Database Query Optimization

**Check for**:
- [ ] **N+1 query problem**: Are related records loaded efficiently?
- [ ] **Missing indexes**: Are frequently queried columns indexed?
- [ ] **Select ***: Are only needed columns selected?
- [ ] **Unnecessary joins**: Are all joins necessary?
- [ ] **Query in loop**: Are queries executed inside loops?
- [ ] **Pagination**: Are large result sets paginated?
- [ ] **Eager loading**: Are associations preloaded when needed?

**Detailed guide**: `references/database-optimization.md` (includes code examples)

---

### ⚡ 2. API Response Optimization

**Check for**:
- [ ] **Response size**: Are responses minimized (exclude unnecessary fields)?
- [ ] **Pagination**: Are large lists paginated?
- [ ] **Filtering**: Are list endpoints filterable to reduce data?
- [ ] **Compression**: Is gzip compression enabled?
- [ ] **Caching headers**: Are appropriate cache headers set?
- [ ] **Partial responses**: Can clients request only needed fields?

**Detailed guide**: `references/api-optimization.md` (includes load testing example)

---

### 🎨 3. Frontend Rendering Optimization

**Check for**:
- [ ] **Unnecessary re-renders**: Are components optimized with `React.memo`/`useMemo`/`useCallback`?
- [ ] **Large lists**: Are large lists virtualized (>100 items)?
- [ ] **Heavy computations**: Are expensive calculations memoized?
- [ ] **Bundle size**: Are large dependencies code-split?
- [ ] **Image optimization**: Are images lazy-loaded and properly sized?
- [ ] **Debouncing/Throttling**: Are frequent events (scroll, input) debounced?

**Detailed guide**: `references/frontend-optimization.md` (includes memoization example)

---

### 💾 4. Caching Strategy

**Check for**:
- [ ] **Cache frequently accessed data**: Are hot paths cached?
- [ ] **Cache invalidation**: Is stale data properly invalidated?
- [ ] **Cache TTL**: Are appropriate expiration times set?
- [ ] **Cache key design**: Are cache keys unique and predictable?
- [ ] **Memory limits**: Is cache size bounded to prevent memory exhaustion?

**Detailed guide**: `references/caching-strategies.md`

---

### 🔄 5. Concurrency & Parallelization

**Check for**:
- [ ] **Blocking operations**: Are long operations executed asynchronously?
- [ ] **Parallel processing**: Can independent tasks run concurrently?
- [ ] **Goroutine leaks**: Are goroutines properly cleaned up?
- [ ] **Race conditions**: Are shared resources protected with mutexes?
- [ ] **Channel buffering**: Are channels sized appropriately?

**Detailed guide**: `references/concurrency-guide.md` (includes goroutine pool pattern)

---

### 📦 6. Memory Management

**Check for**:
- [ ] **Memory leaks**: Are resources (connections, files, goroutines) released?
- [ ] **Large allocations**: Are large objects (slices, maps) pre-allocated when size is known?
- [ ] **Pointer vs value**: Are large structs passed by pointer?
- [ ] **String concatenation**: Is `strings.Builder` used for building large strings?
- [ ] **JSON marshaling**: Are large objects streamed instead of loaded entirely?

**Detailed guide**: `references/memory-management.md`

---

### 🌐 7. Network & HTTP Optimization

**Check for**:
- [ ] **Connection pooling**: Are HTTP clients reused (not created per request)?
- [ ] **Timeouts**: Are reasonable timeouts set on all network requests?
- [ ] **Keep-alive**: Are persistent connections used?
- [ ] **Request batching**: Can multiple requests be combined?
- [ ] **Retry logic**: Are failed requests retried with exponential backoff?

**Detailed guide**: `references/network-optimization.md`

---

### 🔍 8. Algorithm Complexity

**Check for**:
- [ ] **Time complexity**: Is the algorithm O(n²) or worse for large n?
- [ ] **Space complexity**: Does the algorithm use excessive memory?
- [ ] **Redundant work**: Are results computed multiple times?
- [ ] **Early exit**: Can loops exit early when condition is met?

**Detailed guide**: `references/algorithm-complexity.md`

---

## Performance Testing

### Backend Benchmarks
```bash
# Run benchmark tests
go test -bench=. -benchmem ./services/

# Profile CPU and memory
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.
go tool pprof cpu.prof
```

### API Load Testing
```bash
# Apache Bench
ab -n 1000 -c 10 http://localhost:8000/api/v1/nodes

# k6 (see assets/load-test.js)
k6 run assets/load-test.js
```

### Frontend Performance
Run Lighthouse in Chrome DevTools (F12 → Lighthouse tab):
- First Contentful Paint (FCP): < 1.8s
- Largest Contentful Paint (LCP): < 2.5s
- Time to Interactive (TTI): < 3.8s
- Cumulative Layout Shift (CLS): < 0.1

---

## Performance Monitoring

### Backend Instrumentation
```go
import "time"

func (s *Service) CriticalOperation() error {
    start := time.Now()
    defer func() {
        log.Infof("Operation duration: %v", time.Since(start))
    }()
    // ... work ...
}
```

### Database Query Monitoring
Enable slow query logging in development:
```go
db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
    Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
        SlowThreshold: 200 * time.Millisecond, // Log queries > 200ms
        LogLevel:      logger.Warn,
    }),
})
```

---

## Fallback Strategy

If performance issues are unclear:
1. Use profiler to locate actual bottlenecks (`pprof`, Chrome DevTools)
2. Measure before optimizing (benchmarks, load tests)
3. Focus on hot paths (90/10 rule - 90% of time in 10% of code)
4. Don't guess - profile and measure

---

## Exit Criteria

- [ ] All performance issues identified and addressed
- [ ] Benchmark tests pass with acceptable performance
- [ ] No database queries in loops (or justified and documented)
- [ ] Large lists are paginated/virtualized
- [ ] Performance implications documented in PR (if significant)
- [ ] Monitoring/logging added for new critical paths

---

## References

### Codebase
- `services/mihomo/mihomo.go` - Core performance-critical service
- `services/scheduler/speedtest_task.go` - Concurrent node testing example
- `webs/src/views/nodes/` - Frontend list rendering patterns

### External
- Go Performance: https://github.com/dgryski/go-perfbook
- React Performance: https://react.dev/learn/render-and-commit
- GORM Performance: https://gorm.io/docs/performance.html
