import * as monaco from "monaco-editor";

export const getSuggestions = (
  range: monaco.IRange
): monaco.languages.CompletionItem[] => {
  const suggestions: monaco.languages.CompletionItem[] = [
    // 主入口函数
    {
      label: "subMod",
      kind: monaco.languages.CompletionItemKind.Function,
      insertText:
        "function subMod(input, clientType) {\n\t${1:// 修改订阅内容}\n\treturn input;\n}",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "订阅修改脚本的主入口函数。系统会调用此函数来修改订阅内容。",
      detail: "function subMod(input: string, clientType: string): string",
      range: range,
    },
    {
      label: "filterNode",
      kind: monaco.languages.CompletionItemKind.Function,
      insertText:
        "function filterNode(nodes, clientType) {\n\t${1:// 根据条件过滤节点}\n\treturn nodes;\n}",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "节点过滤脚本的主入口函数。系统会调用此函数来过滤节点列表。",
      detail: "function filterNode(nodes: Node[], clientType: string): Node[]",
      range: range,
    },

    // 控制台方法
    {
      label: "console.log",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "console.log(${1:message})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "向服务器控制台输出调试消息。",
      detail: "console.log(...args: any[]): void",
      range: range,
    },
    {
      label: "console.info",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "console.info(${1:message})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "向服务器控制台输出信息消息。",
      detail: "console.info(...args: any[]): void",
      range: range,
    },
    {
      label: "console.warn",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "console.warn(${1:message})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "向服务器控制台输出警告消息。",
      detail: "console.warn(...args: any[]): void",
      range: range,
    },
    {
      label: "console.error",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "console.error(${1:message})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "向服务器控制台输出错误消息。",
      detail: "console.error(...args: any[]): void",
      range: range,
    },

    // ES6+ 特性
    {
      label: "Set",
      kind: monaco.languages.CompletionItemKind.Class,
      insertText: "new Set(${1:iterable})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "创建一个新的 Set 对象，用于存储唯一值。",
      detail: "new Set(iterable?: any[])",
      range: range,
    },
    {
      label: "Map",
      kind: monaco.languages.CompletionItemKind.Class,
      insertText: "new Map(${1:iterable})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "创建一个新的 Map 对象，用于存储键值对。",
      detail: "new Map(iterable?: any[])",
      range: range,
    },

    // 字符串方法
    {
      label: "includes",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "includes(${1:searchString})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "判断一个字符串是否包含另一个字符串，返回 true 或 false。",
      detail:
        "String.prototype.includes(searchString: string, position?: number): boolean",
      range: range,
    },
    {
      label: "startsWith",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "startsWith(${1:searchString})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "判断字符串是否以指定字符串开头，返回 true 或 false。",
      detail:
        "String.prototype.startsWith(searchString: string, position?: number): boolean",
      range: range,
    },
    {
      label: "endsWith",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "endsWith(${1:searchString})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "判断字符串是否以指定字符串结尾，返回 true 或 false。",
      detail:
        "String.prototype.endsWith(searchString: string, length?: number): boolean",
      range: range,
    },
    {
      label: "padStart",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "padStart(${1:targetLength}, ${2:padString})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "用另一个字符串从开头填充当前字符串，直到达到给定长度。",
      detail:
        "String.prototype.padStart(targetLength: number, padString?: string): string",
      range: range,
    },
    {
      label: "padEnd",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "padEnd(${1:targetLength}, ${2:padString})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "用给定字符串从末尾填充当前字符串，直到达到给定长度。",
      detail:
        "String.prototype.padEnd(targetLength: number, padString?: string): string",
      range: range,
    },

    // 数组方法
    {
      label: "find",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "find(${1:callback})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "返回数组中满足测试函数的第一个元素的值，否则返回 undefined。",
      detail:
        "Array.prototype.find(callback: (value: T, index: number, array: T[]) => boolean): T | undefined",
      range: range,
    },
    {
      label: "findIndex",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "findIndex(${1:callback})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "返回数组中满足测试函数的第一个元素的索引，否则返回 -1。",
      detail:
        "Array.prototype.findIndex(callback: (value: T, index: number, array: T[]) => boolean): number",
      range: range,
    },
    {
      label: "filter",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "filter(${1:callback})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "创建一个新数组，包含所有通过测试函数的元素。",
      detail:
        "Array.prototype.filter(callback: (value: T, index: number, array: T[]) => boolean): T[]",
      range: range,
    },
    {
      label: "map",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "map(${1:callback})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "创建一个新数组，包含对原数组每个元素调用提供的函数后的结果。",
      detail:
        "Array.prototype.map(callback: (value: T, index: number, array: T[]) => U): U[]",
      range: range,
    },

    // 对象方法
    {
      label: "Object.assign",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "Object.assign(${1:target}, ${2:source})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "将所有可枚举的自有属性从一个或多个源对象复制到目标对象。返回修改后的目标对象。",
      detail: "Object.assign(target: any, ...sources: any[]): any",
      range: range,
    },
    {
      label: "Object.values",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "Object.values(${1:obj})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "返回对象自有可枚举属性值的数组。",
      detail: "Object.values(obj: any): any[]",
      range: range,
    },
    {
      label: "Object.entries",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "Object.entries(${1:obj})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "返回对象自有可枚举字符串键属性的 [键, 值] 对数组。",
      detail: "Object.entries(obj: any): [string, any][]",
      range: range,
    },
    {
      label: "Object.keys",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "Object.keys(${1:obj})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "返回对象自有可枚举属性名的数组。",
      detail: "Object.keys(obj: any): string[]",
      range: range,
    },

    // Node 接口属性（输入 'node.' 时自动补全）
    {
      label: "ID",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "ID",
      documentation: "节点的唯一标识符",
      detail: "number",
      range: range,
    },
    {
      label: "Link",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Link",
      documentation: "完整链接字符串",
      detail: "string",
      range: range,
    },
    {
      label: "Name",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Name",
      documentation: "节点显示名称",
      detail: "string",
      range: range,
    },
    {
      label: "LinkName",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LinkName",
      documentation: "链接标识名称",
      detail: "string",
      range: range,
    },
    {
      label: "LinkAddress",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LinkAddress",
      documentation: "链接地址",
      detail: "string",
      range: range,
    },
    {
      label: "LinkHost",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LinkHost",
      documentation: "链接主机",
      detail: "string",
      range: range,
    },
    {
      label: "LinkPort",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LinkPort",
      documentation: "链接端口",
      detail: "string",
      range: range,
    },
    {
      label: "DialerProxyName",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "DialerProxyName",
      documentation: "拨号代理名称",
      detail: "string",
      range: range,
    },
    {
      label: "CreateDate",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "CreateDate",
      documentation: "创建日期",
      detail: "string",
      range: range,
    },
    {
      label: "Source",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Source",
      documentation: "节点来源",
      detail: "string",
      range: range,
    },
    {
      label: "SourceID",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "SourceID",
      documentation: "节点来源ID",
      detail: "number",
      range: range,
    },
    {
      label: "Group",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Group",
      documentation: "节点所属分组",
      detail: "string",
      range: range,
    },
    {
      label: "DelayTime",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "DelayTime",
      documentation: "节点延迟（ms）",
      detail: "number",
      range: range,
    },
    {
      label: "Speed",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Speed",
      documentation: "节点速度（MB/s）",
      detail: "number",
      range: range,
    },
    {
      label: "LastCheck",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LastCheck",
      documentation: "最后测试时间",
      detail: "string",
      range: range,
    },
  ];

  return suggestions;
};
