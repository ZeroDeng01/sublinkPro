# Database Query Optimization

This guide covers database performance patterns and anti-patterns for SublinkPro.

## Common Problems

### 1. N+1 Query Problem

**Problem**: Loading related records triggers N additional queries.

**Example violation**:
```go
// ❌ BAD: N+1 query problem
package examples

import "gorm.io/gorm"

type Airport struct {
	ID            uint
	Subscriptions []Subscription
}

type Subscription struct {
	ID        uint
	AirportID uint
}

// N+1 query problem: Triggers 1 + N queries
func GetAirportsWithSubscriptions(db *gorm.DB) ([]Airport, error) {
	airports := []Airport{}
	db.Find(&airports) // 1 query

	for i := range airports {
		// This triggers N additional queries!
		db.Model(&airports[i]).Association("Subscriptions").Find(&airports[i].Subscriptions)
	}

	return airports, nil
}
```

**Solution**: Use `Preload` to eager load associations:
```go
// ✅ GOOD: Eager loading with Preload
package examples

import "gorm.io/gorm"

type Airport struct {
	ID            uint
	Subscriptions []Subscription
}

type Subscription struct {
	ID        uint
	AirportID uint
}

// Eager loading: Executes only 2 queries total
func GetAirportsWithSubscriptions(db *gorm.DB) ([]Airport, error) {
	airports := []Airport{}
	// Preload executes: 1 query for airports + 1 query for all subscriptions
	db.Preload("Subscriptions").Find(&airports)

	return airports, nil
}
```

### 2. Queries Inside Loops

**Problem**: Each loop iteration triggers a database query, resulting in N queries for N items.

**Example violation**:
```go
// ❌ BAD: Query inside loop
for _, nodeID := range nodeIDs {
    var node models.Node
    db.First(&node, nodeID) // N queries!
}
```

**Solution**: Use a single query with an IN clause:
```go
// ✅ GOOD: Single query with IN clause
var nodes []models.Node
db.Where("id IN ?", nodeIDs).Find(&nodes)
```

### 3. Missing Pagination

**Problem**: Fetching large datasets without pagination can cause memory issues and slow response times.

**Example violation**:
```go
// ❌ BAD: No pagination for large datasets
var nodes []models.Node
db.Find(&nodes) // May return 10,000+ records!
```

**Solution**: Implement pagination:
```go
// ✅ GOOD: Pagination
var nodes []models.Node
page := c.Query("page", "1")
pageSize := 50
offset := (page - 1) * pageSize
db.Limit(pageSize).Offset(offset).Find(&nodes)
```

### 4. Missing Indexes

**Problem**: Queries on unindexed columns require full table scans.

**Solution**: Add indexes for frequently queried columns:
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

## Checklist

When reviewing database code:
- [ ] **N+1 query problem**: Are related records loaded efficiently with `Preload`?
- [ ] **Missing indexes**: Are frequently queried columns indexed?
- [ ] **Select ***: Are only needed columns selected?
- [ ] **Unnecessary joins**: Are all joins necessary?
- [ ] **Query in loop**: Are queries executed inside loops?
- [ ] **Pagination**: Are large result sets paginated?
- [ ] **Eager loading**: Are associations preloaded when needed?

## Monitoring

Enable slow query logging in development:
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

## References

- `services/mihomo/mihomo.go` - Core performance-critical service
- GORM Performance: https://gorm.io/docs/performance.html
