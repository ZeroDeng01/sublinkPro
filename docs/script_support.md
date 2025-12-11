# 脚本执行环境支持

脚本执行环境基于 [Goja](https://github.com/dop251/goja)，这是一个纯 Go 语言实现的 ECMAScript 5.1 引擎，包含许多 ES6+ 特性。

为了确保广泛的兼容性，我们为常见的 ES6+ 函数注入了 polyfill。

## 支持的功能

### ES6+ 特性 (原生支持)

Goja 原生支持许多现代 JavaScript 特性，包括但不限于：

- **Let / Const**: 块级作用域变量声明。
- **Arrow Functions**: `(x) => x * 2`。
- **Classes**: `class MyClass { ... }`。
- **Map / Set**: 集合类型。
- **WeakMap / WeakSet**: 弱引用集合。
- **Promise**: 异步编程（注意：脚本执行通常是同步的，但在某些场景下可用）。
- **Symbol**: 唯一标识符。
- **Proxy / Reflect**: 元编程能力。
- **Template Literals**: 模板字符串 \`Hello ${name}\`。
- **Destructuring**: 解构赋值 `const { a, b } = obj`。
- **Default Parameters**: 默认参数 `function(a = 1) { ... }`。
- **Rest / Spread**: 剩余参数和扩展运算符 `...args`。

### 标准库扩展 (Polyfills)

为了方便使用，我们注入了以下 Polyfills：

#### String

- `String.prototype.includes(searchString, position)`
- `String.prototype.startsWith(searchString, position)`
- `String.prototype.endsWith(searchString, position)`
- `String.prototype.padStart(targetLength, padString)`
- `String.prototype.padEnd(targetLength, padString)`
- ...以及标准 ES5 方法。

#### Array

- `Array.from(arrayLike, mapFn, thisArg)`
- `Array.prototype.find(callback)`
- `Array.prototype.findIndex(callback)`
- `Array.prototype.includes(searchElement, fromIndex)`
- ...以及标准 ES5 方法。

#### Object

- `Object.assign(target, ...sources)`
- `Object.values(obj)`
- `Object.entries(obj)`
- ...以及标准 ES5 方法。

### 注入的对象

#### console

我们提供了一个 `console` 对象用于日志记录，输出到服务器日志。

- `console.log(message)`
- `console.info(message)`
- `console.warn(message)`
- `console.error(message)`

## 脚本示例

### 使用 Set 去重

```javascript
function subMod(input, clientType) {
    // 假设 input 是一个逗号分隔的列表
    let items = input.split(',');
    
    // 使用 Set 去重
    let uniqueItems = new Set(items);
    
    // 转回数组并连接
    return Array.from(uniqueItems).join(',');
}
```

### 根据 LinkAddress 去重

```javascript
function filterNode(nodes, clientType) {
    // 使用 Set 存储已存在的 LinkAddress
    const seen = new Set();
    
    return nodes.filter(node => {
        // 如果 LinkAddress 已经存在，则过滤掉
        if (seen.has(node.LinkAddress)) {
            return false;
        }
        // 否则添加到 Set 中并保留
        seen.add(node.LinkAddress);
        return true;
    });
}
```

### 使用 Map 存储键值对

```javascript
function filterNode(nodes, clientType) {
    // nodes: 节点列表数据结构如下
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
    // 使用 Map 统计每个组的节点数量
    let groupCounts = new Map();
    
    nodes.forEach(node => {
        let count = groupCounts.get(node.Group) || 0;
        groupCounts.set(node.Group, count + 1);
    });
    
    // 打印统计信息
    for (let [group, count] of groupCounts) {
        console.log(`Group ${group}: ${count} nodes`);
    }
    
    return nodes;
}
```

### 使用 RegExp 正则匹配

```javascript
function filterNode(nodes, clientType) {
    // 过滤掉名字中包含 "测试" 或 "过期" 的节点（忽略大小写）
    const regex = /(测试|过期)/i;
    
    return nodes.filter(node => !regex.test(node.Name));
}
```

### 使用 Object.entries 遍历对象

```javascript
function subMod(input, clientType) {
    // 假设 input 是 JSON 字符串
    try {
        let config = JSON.parse(input);
        
        // 遍历配置项并修改
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

## 脚本入口点

### 订阅处理脚本

用于修改最终的订阅内容。

```javascript
/**
 * @param {string} input - 原始订阅内容（base64 解码或原始内容）。
 * @param {string} clientType - 客户端类型（例如："v2ray"、"clash"、"surge"）。
 * @returns {string} - 修改后的内容。
 */
function subMod(input, clientType) {
    // 你的逻辑在这里
    return input;
}
```

### 节点过滤脚本

用于在生成订阅之前过滤节点列表。

```javascript
/**
 * @param {Array} nodes - 节点对象数组。
 * @param {string} clientType - 客户端类型。
 * @returns {Array} - 过滤后的节点数组。
 */
function filterNode(nodes, clientType) {
    // 你的逻辑在这里
    return nodes.filter(node => node.remarks.includes("US"));
}
```

## 故障排除

### "TypeError: Cannot read property 'indexOf' of undefined or null"

此错误通常发生在你试图对一个为 `null` 或 `undefined` 的变量调用方法时。
在访问属性之前检查你的数据：

```javascript
if (str && str.indexOf("something") !== -1) {
    // ...
}
```
