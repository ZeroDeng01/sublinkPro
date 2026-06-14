# Algorithm Complexity

This guide covers algorithm optimization patterns for SublinkPro.

## Key Principles

1. Avoid O(n²) algorithms for large datasets
2. Use appropriate data structures (maps for lookups, sets for uniqueness)
3. Eliminate redundant work
4. Exit loops early when condition is met
5. Consider space-time tradeoffs

## Time Complexity

### Problem: O(n²) Nested Loops

**Example violation**:
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
```

### Solution: O(n) with Map

```go
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

## Early Exit

### Problem: Unnecessary Iteration

```go
// ❌ BAD: Continues after finding result
func FindNode(nodes []Node, id uint) *Node {
    var result *Node
    for _, node := range nodes {
        if node.ID == id {
            result = &node
        }
    }
    return result
}
```

### Solution: Early Return

```go
// ✅ GOOD: Exit early
func FindNode(nodes []Node, id uint) *Node {
    for _, node := range nodes {
        if node.ID == id {
            return &node
        }
    }
    return nil
}
```

## Redundant Work

### Problem: Repeated Computation

```go
// ❌ BAD: Recomputes same value
func ProcessNodes(nodes []Node) {
    for _, node := range nodes {
        config := GenerateConfig() // Same for all nodes!
        ApplyConfig(node, config)
    }
}
```

### Solution: Compute Once

```go
// ✅ GOOD: Compute once, reuse
func ProcessNodes(nodes []Node) {
    config := GenerateConfig()
    for _, node := range nodes {
        ApplyConfig(node, config)
    }
}
```

## Data Structure Selection

### Use Maps for Fast Lookups

```go
// ✅ GOOD: O(1) lookup with map
nodeMap := make(map[uint]*Node)
for _, node := range nodes {
    nodeMap[node.ID] = &node
}

// Fast lookup
if node, exists := nodeMap[targetID]; exists {
    // Found in O(1)
}
```

### Use Sets for Uniqueness

```go
// ✅ GOOD: Set using map[T]bool
seen := make(map[string]bool)
unique := []string{}

for _, item := range items {
    if !seen[item] {
        seen[item] = true
        unique = append(unique, item)
    }
}
```

## Space-Time Tradeoffs

Sometimes using more memory improves performance:

```go
// Cache expensive computation results
cache := make(map[string]Result)

func ExpensiveOperation(key string) Result {
    if result, cached := cache[key]; cached {
        return result
    }
    
    result := compute(key) // Expensive
    cache[key] = result
    return result
}
```

## Checklist

When reviewing algorithms:
- [ ] **Time complexity**: Is the algorithm O(n²) or worse for large n?
- [ ] **Space complexity**: Does the algorithm use excessive memory?
- [ ] **Redundant work**: Are results computed multiple times?
- [ ] **Early exit**: Can loops exit early when condition is met?
- [ ] **Data structures**: Are appropriate data structures used (maps for lookups, sets for uniqueness)?

## Common Complexity Classes

| Notation | Name | Example |
|----------|------|---------|
| O(1) | Constant | Map lookup, array access |
| O(log n) | Logarithmic | Binary search |
| O(n) | Linear | Single loop over data |
| O(n log n) | Linearithmic | Efficient sorting |
| O(n²) | Quadratic | Nested loops |
| O(2ⁿ) | Exponential | Recursive fibonacci (naive) |

## Benchmarking

Measure actual performance:
```go
func BenchmarkFindDuplicates(b *testing.B) {
    nodes := generateTestNodes(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        FindDuplicates(nodes)
    }
}

// Run: go test -bench=. -benchmem
```

## References

- Go Performance: https://github.com/dgryski/go-perfbook
