# Script Execution Environment Support

The script execution environment is based on [Goja](https://github.com/dop251/goja), which is an implementation of ECMAScript 5.1 in pure Go with some ES6+ features.

To ensure broad compatibility, we have injected polyfills for common ES6+ functions.

## Supported Features

### Standard Library (ES5.1 + Polyfills)

The following "regular" functions and methods are supported:

#### String
- `String.prototype.includes(searchString, position)`
- `String.prototype.startsWith(searchString, position)`
- `String.prototype.endsWith(searchString, position)`
- `String.prototype.indexOf(searchValue, fromIndex)`
- `String.prototype.lastIndexOf(searchValue, fromIndex)`
- `String.prototype.match(regexp)`
- `String.prototype.replace(regexp|substr, newSubstr|function)`
- `String.prototype.split(separator, limit)`
- `String.prototype.toLowerCase()`
- `String.prototype.toUpperCase()`
- `String.prototype.trim()`
- ...and other standard ES5 String methods.

#### Array
- `Array.prototype.find(callback)`
- `Array.prototype.findIndex(callback)`
- `Array.prototype.includes(searchElement, fromIndex)`
- `Array.prototype.forEach(callback)`
- `Array.prototype.map(callback)`
- `Array.prototype.filter(callback)`
- `Array.prototype.reduce(callback, initialValue)`
- `Array.prototype.push()`, `pop()`, `shift()`, `unshift()`
- `Array.prototype.slice()`, `splice()`
- `Array.prototype.join()`
- ...and other standard ES5 Array methods.

#### Object
- `Object.assign(target, ...sources)`
- `Object.keys(obj)`
- `Object.create(proto)`
- `Object.defineProperty(obj, prop, descriptor)`
- ...and other standard ES5 Object methods.

#### JSON
- `JSON.parse(text)`
- `JSON.stringify(value)`

### Injected Objects

#### console
We provide a `console` object for logging, which outputs to the server logs.
- `console.log(message)`
- `console.info(message)`
- `console.warn(message)`
- `console.error(message)`

## Script Entry Points

### Subscription Processing Script
Used to modify the final subscription content.

```javascript
/**
 * @param {string} content - The original subscription content (base64 decoded or raw).
 * @param {string} clientType - The client type (e.g., "v2ray", "clash", "surge").
 * @returns {string} - The modified content.
 */
function subMod(content, clientType) {
    // Your logic here
    return content;
}
```

### Node Filter Script
Used to filter the list of nodes before generating the subscription.

```javascript
/**
 * @param {Array} nodes - Array of node objects.
 * @param {string} clientType - The client type.
 * @returns {Array} - The filtered array of nodes.
 */
function filterNode(nodes, clientType) {
    // Your logic here
    return nodes.filter(node => node.remarks.includes("US"));
}
```

## Troubleshooting

### "TypeError: Cannot read property 'indexOf' of undefined or null"
This error usually happens if you try to call a method on a variable that is `null` or `undefined`.
Check your data before accessing properties:

```javascript
if (str && str.indexOf("something") !== -1) {
    // ...
}
```
