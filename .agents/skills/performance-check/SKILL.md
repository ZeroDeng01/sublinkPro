# Performance Check Skill

## Purpose

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

**Example violations**:
```go
// ❌ BAD: N+1 query problem
airports := []models.Airport{}
db.Find(&airports)
for _, airport := range airports {
    // This triggers N additional queries!
    db.Model(&airport).Association("Subscriptions").Find(&airport.Subscriptions)
}

// ✅ GOOD: Eager loading with Preload
airports := []models.Airport{}
db.Preload("Subscriptions").Find(&airports)
```

```go
// ❌ BAD: Query inside loop
for _, nodeID := range nodeIDs {
    var node models.Node
    db.First(&node, nodeID) // N queries!
}

// ✅ GOOD: Single query with IN clause
var nodes []models.Node
db.Where("id IN ?", nodeIDs).Find(&nodes)
```

```go
// ❌ BAD: No pagination for large datasets
var nodes []models.Node
db.Find(&nodes) // May return 10,000+ records!

// ✅ GOOD: Pagination
var nodes []models.Node
page := c.Query("page", "1")
pageSize := 50
offset := (page - 1) * pageSize
db.Limit(pageSize).Offset(offset).Find(&nodes)
```

**Database indexing**:
```go
// Add indexes for frequently queried columns
type Node struct {
    ID         uint   `gorm:"primaryKey"`
    AirportID  uint   `gorm:"index"` // ✅ Indexed for foreign key
    Tag        string `gorm:"index"` // ✅ Indexed for filtering
    Name       string // No index needed if rarely queried alone
    CreatedAt  time.Time `gorm:"index"` // ✅ Indexed for sorting
}

// Composite index for multi-column queries
// In migration:
db.Exec("CREATE INDEX idx_airport_tag ON nodes(airport_id, tag)")
```

---

### ⚡ 2. API Response Optimization

**Check for**:
- [ ] **Response size**: Are responses minimized (exclude unnecessary fields)?
- [ ] **Pagination**: Are large lists paginated?
- [ ] **Filtering**: Are list endpoints filterable to reduce data?
- [ ] **Compression**: Is gzip compression enabled?
- [ ] **Caching headers**: Are appropriate cache headers set?
- [ ] **Partial responses**: Can clients request only needed fields?

**Example improvements**:
```go
// ❌ BAD: Returning full objects with sensitive data
func ListNodes(c *gin.Context) {
    var nodes []models.Node
    db.Preload("Airport").Find(&nodes)
    c.JSON(200, nodes) // Includes all fields, even internal ones
}

// ✅ GOOD: DTO with only needed fields, paginated
type NodeListDTO struct {
    ID        uint   `json:"id"`
    Name      string `json:"name"`
    Type      string `json:"type"`
    Location  string `json:"location"`
}

func ListNodes(c *gin.Context) {
    page := getPage(c)
    pageSize := 50
    
    var nodes []models.Node
    var total int64
    
    query := db.Model(&models.Node{})
    query.Count(&total)
    query.Limit(pageSize).Offset((page - 1) * pageSize).Find(&nodes)
    
    dtos := make([]NodeListDTO, len(nodes))
    for i, node := range nodes {
        dtos[i] = NodeListDTO{
            ID:       node.ID,
            Name:     node.Name,
            Type:     node.Type,
            Location: node.Location,
        }
    }
    
    c.JSON(200, gin.H{
        "data":     dtos,
        "total":    total,
        "page":     page,
        "pageSize": pageSize,
    })
}
```

**Enable compression**:
```go
import "github.com/gin-contrib/gzip"

router.Use(gzip.Gzip(gzip.DefaultCompression))
```

---

### 🎨 3. Frontend Rendering Optimization

**Check for**:
- [ ] **Unnecessary re-renders**: Are components optimized with `React.memo`/`useMemo`/`useCallback`?
- [ ] **Large lists**: Are large lists virtualized?
- [ ] **Heavy computations**: Are expensive calculations memoized?
- [ ] **Bundle size**: Are large dependencies code-split?
- [ ] **Image optimization**: Are images lazy-loaded and properly sized?
- [ ] **Debouncing/Throttling**: Are frequent events (scroll, input) debounced?

**Example violations**:
```jsx
// ❌ BAD: Component re-renders on every parent render
function NodeItem({ node }) {
    const formattedDate = formatDate(node.createdAt); // Computed every render!
    return <div>{node.name} - {formattedDate}</div>;
}

// ✅ GOOD: Memoized component and computation
const NodeItem = React.memo(({ node }) => {
    const formattedDate = useMemo(
        () => formatDate(node.createdAt),
        [node.createdAt]
    );
    return <div>{node.name} - {formattedDate}</div>;
});
```

```jsx
// ❌ BAD: Rendering 1000+ items without virtualization
function NodeList({ nodes }) {
    return (
        <div>
            {nodes.map(node => <NodeItem key={node.id} node={node} />)}
        </div>
    );
}

// ✅ GOOD: Virtualized list for large datasets
import { FixedSizeList } from 'react-window';

function NodeList({ nodes }) {
    return (
        <FixedSizeList
            height={600}
            itemCount={nodes.length}
            itemSize={50}
            width="100%"
        >
            {({ index, style }) => (
                <div style={style}>
                    <NodeItem node={nodes[index]} />
                </div>
            )}
        </FixedSizeList>
    );
}
```

```jsx
// ❌ BAD: Search triggers API call on every keystroke
function SearchInput() {
    const [query, setQuery] = useState('');
    
    const handleChange = (e) => {
        setQuery(e.target.value);
        fetchResults(e.target.value); // API call on every keystroke!
    };
    
    return <input value={query} onChange={handleChange} />;
}

// ✅ GOOD: Debounced search
import { debounce } from 'lodash-es';

function SearchInput() {
    const [query, setQuery] = useState('');
    
    const debouncedFetch = useMemo(
        () => debounce((q) => fetchResults(q), 300),
        []
    );
    
    const handleChange = (e) => {
        setQuery(e.target.value);
        debouncedFetch(e.target.value);
    };
    
    return <input value={query} onChange={handleChange} />;
}
```

---

### 💾 4. Caching Strategy

**Check for**:
- [ ] **Cache frequently accessed data**: Are hot paths cached?
- [ ] **Cache invalidation**: Is stale data properly invalidated?
- [ ] **Cache TTL**: Are appropriate expiration times set?
- [ ] **Cache key design**: Are cache keys unique and predictable?
- [ ] **Memory limits**: Is cache size bounded to prevent memory exhaustion?

**Backend caching example**:
```go
// ✅ GOOD: Cache mihomo config generation
var configCache = cache.New(5*time.Minute, 10*time.Minute)

func GenerateMihomoConfig(clientID uint) (string, error) {
    cacheKey := fmt.Sprintf("config:%d", clientID)
    
    // Check cache first
    if cached, found := configCache.Get(cacheKey); found {
        return cached.(string), nil
    }
    
    // Generate config (expensive operation)
    config, err := buildMihomoConfig(clientID)
    if err != nil {
        return "", err
    }
    
    // Store in cache
    configCache.Set(cacheKey, config, cache.DefaultExpiration)
    return config, nil
}

// Invalidate cache when nodes change
func UpdateNode(node *models.Node) error {
    if err := db.Save(node).Error; err != nil {
        return err
    }
    
    // Invalidate affected client configs
    configCache.Delete(fmt.Sprintf("config:%d", node.ClientID))
    return nil
}
```

**Frontend caching with SWR**:
```jsx
// ✅ GOOD: SWR provides automatic caching and revalidation
import useSWR from 'swr';

function NodeList() {
    const { data: nodes, error, mutate } = useSWR('/api/v1/nodes', fetcher, {
        revalidateOnFocus: false,
        dedupingInterval: 60000, // Dedupe requests within 1 minute
    });
    
    if (error) return <div>Failed to load</div>;
    if (!nodes) return <div>Loading...</div>;
    
    return <div>{/* render nodes */}</div>;
}
```

---

### 🔄 5. Concurrency & Parallelization

**Check for**:
- [ ] **Blocking operations**: Are long operations executed asynchronously?
- [ ] **Parallel processing**: Can independent tasks run concurrently?
- [ ] **Goroutine leaks**: Are goroutines properly cleaned up?
- [ ] **Race conditions**: Are shared resources protected with mutexes?
- [ ] **Channel buffering**: Are channels sized appropriately?

**Example improvements**:
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

// ✅ GOOD: Parallel processing with goroutines
func TestMultipleNodes(nodeIDs []uint) []TestResult {
    results := make([]TestResult, len(nodeIDs))
    var wg sync.WaitGroup
    
    // Limit concurrency to avoid overwhelming resources
    semaphore := make(chan struct{}, 10) // Max 10 concurrent
    
    for i, id := range nodeIDs {
        wg.Add(1)
        go func(index int, nodeID uint) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            results[index] = testNode(nodeID)
        }(i, id)
    }
    
    wg.Wait()
    return results
}
```

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

---

### 📦 6. Memory Management

**Check for**:
- [ ] **Memory leaks**: Are resources (connections, files, goroutines) released?
- [ ] **Large allocations**: Are large objects (slices, maps) pre-allocated when size is known?
- [ ] **Pointer vs value**: Are large structs passed by pointer?
- [ ] **String concatenation**: Is `strings.Builder` used for building large strings?
- [ ] **JSON marshaling**: Are large objects streamed instead of loaded entirely?

**Example improvements**:
```go
// ❌ BAD: Repeated allocations in loop
var result string
for _, item := range items {
    result += item + "\n" // Creates new string each iteration!
}

// ✅ GOOD: strings.Builder
var builder strings.Builder
builder.Grow(len(items) * 50) // Pre-allocate if size is known
for _, item := range items {
    builder.WriteString(item)
    builder.WriteString("\n")
}
result := builder.String()
```

```go
// ❌ BAD: Loading large file entirely into memory
data, err := os.ReadFile("large_file.json") // May be 100MB+
json.Unmarshal(data, &result)

// ✅ GOOD: Streaming parser
file, err := os.Open("large_file.json")
defer file.Close()
decoder := json.NewDecoder(file)
decoder.Decode(&result) // Processes incrementally
```

---

### 🌐 7. Network & HTTP Optimization

**Check for**:
- [ ] **Connection pooling**: Are HTTP clients reused (not created per request)?
- [ ] **Timeouts**: Are reasonable timeouts set on all network requests?
- [ ] **Keep-alive**: Are persistent connections used?
- [ ] **Request batching**: Can multiple requests be combined?
- [ ] **Retry logic**: Are failed requests retried with exponential backoff?

**Example improvements**:
```go
// ❌ BAD: Creating new client for each request
func FetchSubscription(url string) ([]byte, error) {
    client := &http.Client{} // New client each time!
    resp, err := client.Get(url)
    // ...
}

// ✅ GOOD: Shared HTTP client with connection pool
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

func FetchSubscription(url string) ([]byte, error) {
    resp, err := httpClient.Get(url)
    // ...
}
```

---

### 🔍 8. Algorithm Complexity

**Check for**:
- [ ] **Time complexity**: Is the algorithm O(n²) or worse for large n?
- [ ] **Space complexity**: Does the algorithm use excessive memory?
- [ ] **Redundant work**: Are results computed multiple times?
- [ ] **Early exit**: Can loops exit early when condition is met?

**Example improvements**:
```go
// ❌ BAD: O(n²) nested loops
func FindDuplicates(nodes []Node) []Node {
    duplicates := []Node{}
    for i, node1 := range nodes {
        for j, node2 := range nodes {
            if i != j && node1.Name == node2.Name {
                duplicates = append(duplicates, node1)
            }
        }
    }
    return duplicates
}

// ✅ GOOD: O(n) with map
func FindDuplicates(nodes []Node) []Node {
    seen := make(map[string]bool)
    duplicates := []Node{}
    
    for _, node := range nodes {
        if seen[node.Name] {
            duplicates = append(duplicates, node)
        } else {
            seen[node.Name] = true
        }
    }
    return duplicates
}
```

---

## Performance Testing

### 1. Benchmark Tests (Backend)

**Create benchmark tests for performance-critical code**:
```go
// services/node_service_test.go
func BenchmarkGenerateConfig(b *testing.B) {
    // Setup
    nodes := generateTestNodes(1000)
    
    b.ResetTimer() // Start timing
    for i := 0; i < b.N; i++ {
        GenerateConfig(nodes)
    }
}

// Run: go test -bench=. -benchmem ./services/
```

**Interpret results**:
- `ns/op`: Nanoseconds per operation (lower is better)
- `B/op`: Bytes allocated per operation (lower is better)
- `allocs/op`: Allocations per operation (lower is better)

### 2. Load Testing (API)

**Use tools to simulate load**:
```bash
# Apache Bench: Simple load test
ab -n 1000 -c 10 http://localhost:8000/api/v1/nodes

# k6: Advanced load testing
k6 run load-test.js
```

**Example k6 script**:
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    vus: 10, // 10 virtual users
    duration: '30s',
};

export default function () {
    let res = http.get('http://localhost:8000/api/v1/nodes');
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });
    sleep(1);
}
```

### 3. Frontend Performance (Lighthouse)

**Run Lighthouse in Chrome DevTools**:
1. Open DevTools (F12)
2. Go to "Lighthouse" tab
3. Run analysis
4. Review metrics: FCP, LCP, TTI, CLS

**Performance budget**:
- First Contentful Paint (FCP): < 1.8s
- Largest Contentful Paint (LCP): < 2.5s
- Time to Interactive (TTI): < 3.8s
- Cumulative Layout Shift (CLS): < 0.1

---

## Performance Monitoring

### Backend Metrics

**Add instrumentation for key operations**:
```go
import "time"

func (s *NodeService) GenerateConfig(clientID uint) (string, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        log.Infof("GenerateConfig duration: %v", duration)
        // Send to monitoring system (Prometheus, etc.)
    }()
    
    // ... actual work ...
}
```

### Database Query Monitoring

**Enable slow query logging** (in development):
```go
// config/database.go
db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // Log all queries
})

// Or custom logger for slow queries only
db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
    Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
        SlowThreshold: 200 * time.Millisecond, // Log queries > 200ms
        LogLevel:      logger.Warn,
    }),
})
```

---

## Performance Checklist Summary

Before completing performance review:
- [ ] No N+1 query problems (use Preload)
- [ ] Large result sets are paginated
- [ ] Appropriate database indexes exist
- [ ] API responses are minimized (DTOs, not full models)
- [ ] Frontend lists are virtualized if > 100 items
- [ ] Expensive computations are memoized
- [ ] Frequent events (search, scroll) are debounced
- [ ] Hot paths are cached with proper invalidation
- [ ] Independent tasks run concurrently when possible
- [ ] HTTP clients are reused (connection pooling)
- [ ] No O(n²) algorithms for large datasets
- [ ] Benchmark tests added for critical paths
- [ ] Load testing performed for new API endpoints
- [ ] Lighthouse score > 90 for new frontend pages

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

- `services/mihomo/mihomo.go` - Core performance-critical service
- `services/scheduler/speedtest_task.go` - Concurrent node testing example
- `webs/src/views/nodes/` - Frontend list rendering patterns
- Go Performance: https://github.com/dgryski/go-perfbook
- React Performance: https://react.dev/learn/render-and-commit
- GORM Performance: https://gorm.io/docs/performance.html
