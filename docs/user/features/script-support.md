English | [简体中文](script-support.zh-CN.md)

[Guide index](../README.md)

# Script Runtime Support

The script runtime is based on [Goja](https://github.com/dop251/goja), a pure Go ECMAScript 5.1 engine with many ES6+ features.

To improve compatibility, common ES6+ functions are injected as polyfills.

## Supported Features

### ES6+ features, natively supported

Goja supports many modern JavaScript features, including but not limited to:

- **Let / Const**: block scoped variable declarations.
- **Arrow Functions**: `(x) => x * 2`.
- **Classes**: `class MyClass { ... }`.
- **Map / Set**: collection types.
- **WeakMap / WeakSet**: weak reference collections.
- **Promise**: asynchronous programming. Script execution is usually synchronous, but this is available in some scenarios.
- **Symbol**: unique identifiers.
- **Proxy / Reflect**: metaprogramming capabilities.
- **Template Literals**: template strings, `Hello ${name}`.
- **Destructuring**: `const { a, b } = obj`.
- **Default Parameters**: `function(a = 1) { ... }`.
- **Rest / Spread**: rest parameters and spread operator, `...args`.

### Standard library extensions, polyfills

The following polyfills are injected for convenience:

#### String

- `String.prototype.includes(searchString, position)`
- `String.prototype.startsWith(searchString, position)`
- `String.prototype.endsWith(searchString, position)`
- `String.prototype.padStart(targetLength, padString)`
- `String.prototype.padEnd(targetLength, padString)`
- Standard ES5 methods are also available.

#### Array

- `Array.from(arrayLike, mapFn, thisArg)`
- `Array.prototype.find(callback)`
- `Array.prototype.findIndex(callback)`
- `Array.prototype.includes(searchElement, fromIndex)`
- Standard ES5 methods are also available.

#### Object

- `Object.assign(target, ...sources)`
- `Object.values(obj)`
- `Object.entries(obj)`
- Standard ES5 methods are also available.

### Injected objects

#### console

A `console` object is provided for logging to server logs.

- `console.log(message)`
- `console.info(message)`
- `console.warn(message)`
- `console.error(message)`

## Script Examples

### Deduplicate with Set

```javascript
function subMod(input, clientType) {
    // Assume input is a comma separated list
    let items = input.split(',');
    
    // Deduplicate with Set
    let uniqueItems = new Set(items);
    
    // Convert back to array and join
    return Array.from(uniqueItems).join(',');
}
```

### Deduplicate by LinkAddress

```javascript
function filterNode(nodes, clientType) {
    // Store seen LinkAddress values in Set
    const seen = new Set();
    
    return nodes.filter(node => {
        // Filter out duplicated LinkAddress
        if (seen.has(node.LinkAddress)) {
            return false;
        }
        // Otherwise add it and keep the node
        seen.add(node.LinkAddress);
        return true;
    });
}
```

### Use Map for key value storage

```javascript
function filterNode(nodes, clientType) {
    // nodes: node list data structure is shown below
    // [
    //     {
    //         "ID": 1,
    //         "Link": "vmess://4564564646",
    //         "Name": "xx订阅_US-CDN-SSL",
    //         "LinkName": "US-CDN-SSL",
    //         "LinkAddress": "xxxxxxxxx.net:443",
    //         "LinkHost": "xxxxxxxxx.net",
    //         "LinkPort": "443",
    //         "DialerProxyName": "",
    //         "CreateDate": "",
    //         "Source": "manual",
    //         "SourceID": 0,
    //         "Group": "自用",
    //         "Speed": 110,
    //         "LatencyCheckAt": "2025-11-26 23:49:58",
    //         "SpeedCheckAt": "2025-11-26 23:50:15"
    //     }, {
    //         "ID": 2,
    //         "Link": "vmess://456456464611111",
    //         "Name": "xx订阅_US-CDN-SSL1",
    //         "LinkName": "US-CDN-SSL1",
    //         "LinkAddress": "xxxxxxxxx1.net:443",
    //         "LinkHost": "xxxxxxxxx1.net",
    //         "LinkPort": "443",
    //         "DialerProxyName": "",
    //         "CreateDate": "",
    //         "Source": "manual",
    //         "SourceID": 0,
    //         "Group": "自用",
    //         "Speed": 100,
    //         "LatencyCheckAt": "2025-11-26 23:49:58",
    //         "SpeedCheckAt": "2025-11-26 23:50:20"
    //     }
    // ]
    // Use Map to count nodes in each group
    let groupCounts = new Map();
    
    nodes.forEach(node => {
        let count = groupCounts.get(node.Group) || 0;
        groupCounts.set(node.Group, count + 1);
    });
    
    // Print statistics
    for (let [group, count] of groupCounts) {
        console.log(`Group ${group}: ${count} nodes`);
    }
    
    return nodes;
}
```

### Match with RegExp

```javascript
function filterNode(nodes, clientType) {
    // Filter nodes whose names contain "测试" or "过期", case insensitive
    const regex = /(测试|过期)/i;
    
    return nodes.filter(node => !regex.test(node.Name));
}
```

### Iterate objects with Object.entries

```javascript
function subMod(input, clientType) {
    // Assume input is a JSON string
    try {
        let config = JSON.parse(input);
        
        // Iterate and modify config entries
        for (let [key, value] of Object.entries(config)) {
            if (typeof value === 'string' && value.includes('old-domain.com')) {
                config[key] = value.replace('old-domain.com', 'new-domain.com');
            }
        }
        
        return JSON.stringify(config, null, 2);
    } catch (e) {
        console.error("Parse error: " + e);
        return input;
    }
}
```

## Script Entry Points

### Subscription processing script

Used to modify the final subscription content.

```javascript
/**
 * @param {string} input - Original subscription content, base64 decoded or raw content.
 * @param {string} clientType - Client type, such as "v2ray", "clash", or "surge".
 * @returns {string} - Modified content.
 */
function subMod(input, clientType) {
    // Your logic here
    return input;
}
```

### Node filtering script

Used to filter the node list before subscription generation.

```javascript
/**
 * @param {Array} nodes - Array of node objects.
 * @param {string} clientType - Client type.
 * @returns {Array} - Filtered node array.
 */
function filterNode(nodes, clientType) {
    // Your logic here
    return nodes.filter(node => node.remarks.includes("US"));
}
```

## Troubleshooting

### "TypeError: Cannot read property 'indexOf' of undefined or null"

This usually happens when you call a method on a variable that is `null` or `undefined`.
Check data before accessing properties:

```javascript
if (str && str.indexOf("something") !== -1) {
    // ...
}
```
