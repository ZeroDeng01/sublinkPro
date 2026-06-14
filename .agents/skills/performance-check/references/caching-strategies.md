# Caching Strategies

This guide covers caching patterns for SublinkPro backend and frontend.

## Key Principles

1. Cache frequently accessed data
2. Implement proper cache invalidation
3. Set appropriate TTL (Time To Live)
4. Design unique and predictable cache keys
5. Bound cache size to prevent memory exhaustion

## Backend Caching

### In-Memory Cache Example

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

## Cache Invalidation Patterns

### 1. Time-Based (TTL)
```go
cache.Set(key, value, 5*time.Minute) // Expires after 5 minutes
```

### 2. Event-Based
```go
// Invalidate when data changes
func UpdateAirport(airport *models.Airport) error {
    if err := db.Save(airport).Error; err != nil {
        return err
    }
    cache.Delete(fmt.Sprintf("airport:%d", airport.ID))
    return nil
}
```

### 3. Tag-Based
```go
// Invalidate all caches related to a client
func InvalidateClientCaches(clientID uint) {
    cache.DeleteMatching(fmt.Sprintf("client:%d:*", clientID))
}
```

## Frontend Caching

### Using SWR (Stale-While-Revalidate)

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

### Manual Cache Invalidation

```jsx
// Force refresh after mutation
const { mutate } = useSWR('/api/v1/nodes');

const handleUpdate = async () => {
    await updateNode(node);
    mutate(); // Revalidate cache
};
```

## Cache Key Design

### Good Cache Keys
- Unique per data entity: `config:{clientID}`
- Include version: `nodes:v2:{airportID}`
- Hierarchical: `client:{id}:config`, `client:{id}:nodes`

### Poor Cache Keys
- Too generic: `config`, `nodes`
- Not predictable: `random123`
- Missing context: `data`

## Checklist

When reviewing caching code:
- [ ] **Cache frequently accessed data**: Are hot paths cached?
- [ ] **Cache invalidation**: Is stale data properly invalidated?
- [ ] **Cache TTL**: Are appropriate expiration times set?
- [ ] **Cache key design**: Are cache keys unique and predictable?
- [ ] **Memory limits**: Is cache size bounded to prevent memory exhaustion?

## Common Caching Targets

1. **Mihomo configs** - Expensive to generate, valid until nodes change
2. **Subscription responses** - Can cache for 1-5 minutes
3. **Node lists** - Cache with short TTL (30-60s)
4. **DNS resolutions** - Cache with TTL from DNS records
5. **API responses** - Use HTTP cache headers

## Monitoring

Track cache hit rates:
```go
hits := cache.Hits()
misses := cache.Misses()
hitRate := float64(hits) / float64(hits + misses)
log.Infof("Cache hit rate: %.2f%%", hitRate * 100)
```
