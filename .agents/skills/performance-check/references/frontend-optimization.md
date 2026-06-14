# Frontend Rendering Optimization

This guide covers React performance patterns for SublinkPro's frontend.

## Key Principles

1. Minimize unnecessary re-renders
2. Virtualize large lists
3. Memoize expensive computations
4. Debounce/throttle frequent events
5. Code-split large dependencies
6. Lazy-load images

## Preventing Unnecessary Re-renders

### Problem: Component Re-renders on Every Parent Render

**Example violation**:
```jsx
// ❌ BAD: Component re-renders on every parent render
function NodeItem({ node }) {
    const formattedDate = formatDate(node.createdAt); // Computed every render!
    return <div>{node.name} - {formattedDate}</div>;
}
```

### Solution: React.memo + useMemo

```jsx
// ✅ GOOD: React component memoization example

import React, { useMemo } from 'react';

// Memoized component prevents unnecessary re-renders
const NodeItem = React.memo(({ node }) => {
    // Memoize expensive computation
    const formattedDate = useMemo(
        () => formatDate(node.createdAt),
        [node.createdAt]
    );

    return (
        <div>
            {node.name} - {formattedDate}
        </div>
    );
});

function formatDate(timestamp) {
    // Expensive formatting operation
    return new Date(timestamp).toLocaleString();
}

export default NodeItem;
```

## Virtualizing Large Lists

### Problem: Rendering 1000+ Items

**Example violation**:
```jsx
// ❌ BAD: Rendering 1000+ items without virtualization
function NodeList({ nodes }) {
    return (
        <div>
            {nodes.map(node => <NodeItem key={node.id} node={node} />)}
        </div>
    );
}
```

### Solution: react-window

```jsx
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

## Debouncing Events

### Problem: API Calls on Every Keystroke

**Example violation**:
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
```

### Solution: Debounced Search

```jsx
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

## Checklist

When reviewing frontend code:
- [ ] **Unnecessary re-renders**: Are components optimized with `React.memo`/`useMemo`/`useCallback`?
- [ ] **Large lists**: Are large lists virtualized (>100 items)?
- [ ] **Heavy computations**: Are expensive calculations memoized?
- [ ] **Bundle size**: Are large dependencies code-split?
- [ ] **Image optimization**: Are images lazy-loaded and properly sized?
- [ ] **Debouncing/Throttling**: Are frequent events (scroll, input) debounced?

## Performance Metrics (Lighthouse)

Run Lighthouse in Chrome DevTools:
1. Open DevTools (F12)
2. Go to "Lighthouse" tab
3. Run analysis
4. Review metrics: FCP, LCP, TTI, CLS

**Performance budget**:
- First Contentful Paint (FCP): < 1.8s
- Largest Contentful Paint (LCP): < 2.5s
- Time to Interactive (TTI): < 3.8s
- Cumulative Layout Shift (CLS): < 0.1

## References

- `webs/src/views/nodes/` - Frontend list rendering patterns
- React Performance: https://react.dev/learn/render-and-commit
