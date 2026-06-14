# Network & HTTP Optimization

This guide covers network performance patterns for SublinkPro.

## Key Principles

1. Reuse HTTP clients (connection pooling)
2. Set reasonable timeouts on all network requests
3. Use persistent connections (keep-alive)
4. Batch requests when possible
5. Implement retry logic with exponential backoff

## Connection Pooling

### Problem: Creating New Client Per Request

**Example violation**:
```go
// ❌ BAD: Creating new client for each request
func FetchSubscription(url string) ([]byte, error) {
    client := &http.Client{} // New client each time!
    resp, err := client.Get(url)
    // ...
}
```

### Solution: Shared HTTP Client

```go
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

## Timeout Configuration

Always set timeouts to prevent hanging requests:

```go
// ✅ GOOD: Comprehensive timeout configuration
client := &http.Client{
    Timeout: 30 * time.Second, // Overall request timeout
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second, // Connection timeout
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    },
}
```

## Request Batching

When multiple independent requests are needed, execute them concurrently:

```go
// ✅ GOOD: Parallel HTTP requests
func FetchMultipleSubscriptions(urls []string) ([][]byte, error) {
    results := make([][]byte, len(urls))
    var wg sync.WaitGroup
    var mu sync.Mutex
    errors := []error{}
    
    for i, url := range urls {
        wg.Add(1)
        go func(index int, u string) {
            defer wg.Done()
            data, err := FetchSubscription(u)
            if err != nil {
                mu.Lock()
                errors = append(errors, err)
                mu.Unlock()
                return
            }
            results[index] = data
        }(i, url)
    }
    
    wg.Wait()
    if len(errors) > 0 {
        return nil, errors[0]
    }
    return results, nil
}
```

## Retry Logic

Implement exponential backoff for transient failures:

```go
// ✅ GOOD: Retry with exponential backoff
func FetchWithRetry(url string, maxRetries int) ([]byte, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        resp, err := httpClient.Get(url)
        if err == nil {
            defer resp.Body.Close()
            if resp.StatusCode == 200 {
                return io.ReadAll(resp.Body)
            }
            lastErr = fmt.Errorf("status %d", resp.StatusCode)
        } else {
            lastErr = err
        }
        
        // Exponential backoff: 1s, 2s, 4s, 8s...
        if i < maxRetries-1 {
            time.Sleep(time.Duration(1<<i) * time.Second)
        }
    }
    
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

## Checklist

When reviewing network code:
- [ ] **Connection pooling**: Are HTTP clients reused (not created per request)?
- [ ] **Timeouts**: Are reasonable timeouts set on all network requests?
- [ ] **Keep-alive**: Are persistent connections used?
- [ ] **Request batching**: Can multiple requests be combined or parallelized?
- [ ] **Retry logic**: Are failed requests retried with exponential backoff?

## SublinkPro Pattern

SublinkPro uses a shared proxy client in `utils/proxy_client.go`:
- Reuses connections
- Supports proxy chaining
- Configurable timeouts
- Used for subscription fetching and DNS resolution

## References

- `utils/proxy_client.go` - Shared proxy HTTP client
- `node/sub.go` - Subscription fetching with proxy client
