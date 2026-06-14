# API Response Optimization

This guide covers API performance best practices for SublinkPro.

## Key Principles

1. Minimize response payload size
2. Implement pagination for list endpoints
3. Use DTOs to exclude unnecessary fields
4. Enable compression
5. Set appropriate cache headers

## Response Size Optimization

### Problem: Returning Full Objects

**Example violation**:
```go
// ❌ BAD: Returning full objects with sensitive data
func ListNodes(c *gin.Context) {
    var nodes []models.Node
    db.Preload("Airport").Find(&nodes)
    c.JSON(200, nodes) // Includes all fields, even internal ones
}
```

### Solution: Use DTOs

Create Data Transfer Objects that contain only the fields needed by the client:

```go
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

## Enable Compression

Use gzip middleware to compress responses:

```go
import "github.com/gin-contrib/gzip"

router.Use(gzip.Gzip(gzip.DefaultCompression))
```

## Pagination Pattern

Standard pagination response format:
```json
{
  "data": [...],
  "total": 1000,
  "page": 1,
  "pageSize": 50
}
```

## Checklist

When reviewing API endpoints:
- [ ] **Response size**: Are responses minimized (exclude unnecessary fields)?
- [ ] **Pagination**: Are large lists paginated?
- [ ] **Filtering**: Are list endpoints filterable to reduce data?
- [ ] **Compression**: Is gzip compression enabled?
- [ ] **Caching headers**: Are appropriate cache headers set?
- [ ] **Partial responses**: Can clients request only needed fields?

## Load Testing

Use tools to verify API performance:

```bash
# Apache Bench: Simple load test
ab -n 1000 -c 10 http://localhost:8000/api/v1/nodes

# k6: Advanced load testing
k6 run load-test.js
```

Example k6 script:

```javascript
// k6 load test script for SublinkPro API
// Run: k6 run load-test.js

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
