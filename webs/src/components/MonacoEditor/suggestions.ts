import * as monaco from "monaco-editor";

export const getSuggestions = (
  range: monaco.IRange
): monaco.languages.CompletionItem[] => {
  const suggestions: monaco.languages.CompletionItem[] = [
    // Main Entry Functions
    {
      label: "subMod",
      kind: monaco.languages.CompletionItemKind.Function,
      insertText:
        "function subMod(input, clientType) {\n\t${1:// Modify subscription content}\n\treturn input;\n}",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "Main entry point for subscription modification scripts. This function is called by the system to modify subscription content.",
      detail: "function subMod(input: string, clientType: string): string",
      range: range,
    },
    {
      label: "filterNode",
      kind: monaco.languages.CompletionItemKind.Function,
      insertText:
        "function filterNode(nodes, clientType) {\n\t${1:// Filter nodes based on your criteria}\n\treturn nodes;\n}",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "Main entry point for node filtering scripts. This function is called by the system to filter the list of nodes.",
      detail: "function filterNode(nodes: Node[], clientType: string): Node[]",
      range: range,
    },

    // Console Methods
    {
      label: "console.log",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "console.log(${1:message})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "Output a message to the server console for debugging purposes.",
      detail: "console.log(...args: any[]): void",
      range: range,
    },
    {
      label: "console.info",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "console.info(${1:message})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "Output an informational message to the server console.",
      detail: "console.info(...args: any[]): void",
      range: range,
    },
    {
      label: "console.warn",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "console.warn(${1:message})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "Output a warning message to the server console.",
      detail: "console.warn(...args: any[]): void",
      range: range,
    },
    {
      label: "console.error",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "console.error(${1:message})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "Output an error message to the server console.",
      detail: "console.error(...args: any[]): void",
      range: range,
    },

    // ES6+ Features
    {
      label: "Set",
      kind: monaco.languages.CompletionItemKind.Class,
      insertText: "new Set(${1:iterable})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "Create a new Set object for storing unique values.",
      detail: "new Set(iterable?: any[])",
      range: range,
    },
    {
      label: "Map",
      kind: monaco.languages.CompletionItemKind.Class,
      insertText: "new Map(${1:iterable})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: "Create a new Map object for storing key-value pairs.",
      detail: "new Map(iterable?: any[])",
      range: range,
    },

    // String Methods
    {
      label: "includes",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "includes(${1:searchString})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "Determines whether one string may be found within another string, returning true or false as appropriate.",
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
        "Determines whether a string begins with the characters of a specified string, returning true or false as appropriate.",
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
        "Determines whether a string ends with the characters of a specified string, returning true or false as appropriate.",
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
        "Pads the current string with another string until the resulting string reaches the given length. The padding is applied from the start of the current string.",
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
        "Pads the current string with a given string until the resulting string reaches the given length. The padding is applied from the end of the current string.",
      detail:
        "String.prototype.padEnd(targetLength: number, padString?: string): string",
      range: range,
    },

    // Array Methods
    {
      label: "find",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "find(${1:callback})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "Returns the value of the first element in the array that satisfies the provided testing function. Otherwise undefined is returned.",
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
        "Returns the index of the first element in the array that satisfies the provided testing function. Otherwise, -1 is returned.",
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
        "Creates a new array with all elements that pass the test implemented by the provided function.",
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
        "Creates a new array populated with the results of calling a provided function on every element in the calling array.",
      detail:
        "Array.prototype.map(callback: (value: T, index: number, array: T[]) => U): U[]",
      range: range,
    },

    // Object Methods
    {
      label: "Object.assign",
      kind: monaco.languages.CompletionItemKind.Method,
      insertText: "Object.assign(${1:target}, ${2:source})",
      insertTextRules:
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation:
        "Copies all enumerable own properties from one or more source objects to a target object. It returns the modified target object.",
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
        "Returns an array of a given object's own enumerable property values.",
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
        "Returns an array of a given object's own enumerable string-keyed property [key, value] pairs.",
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
        "Returns an array of a given object's own enumerable property names.",
      detail: "Object.keys(obj: any): string[]",
      range: range,
    },

    // Node Interface Properties (for autocomplete when typing 'node.')
    {
      label: "ID",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "ID",
      documentation: "Unique identifier for the node",
      detail: "number",
      range: range,
    },
    {
      label: "Link",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Link",
      documentation: "Full link string",
      detail: "string",
      range: range,
    },
    {
      label: "Name",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Name",
      documentation: "Display name of the node",
      detail: "string",
      range: range,
    },
    {
      label: "LinkName",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LinkName",
      documentation: "Link identifier name",
      detail: "string",
      range: range,
    },
    {
      label: "LinkAddress",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LinkAddress",
      documentation: "Link address",
      detail: "string",
      range: range,
    },
    {
      label: "LinkHost",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LinkHost",
      documentation: "Link host",
      detail: "string",
      range: range,
    },
    {
      label: "LinkPort",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LinkPort",
      documentation: "Link port",
      detail: "string",
      range: range,
    },
    {
      label: "DialerProxyName",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "DialerProxyName",
      documentation: "Dialer proxy name",
      detail: "string",
      range: range,
    },
    {
      label: "CreateDate",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "CreateDate",
      documentation: "Creation date",
      detail: "string",
      range: range,
    },
    {
      label: "Source",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Source",
      documentation: "Source identifier",
      detail: "string",
      range: range,
    },
    {
      label: "SourceID",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "SourceID",
      documentation: "Source ID number",
      detail: "number",
      range: range,
    },
    {
      label: "Group",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Group",
      documentation: "Node group",
      detail: "string",
      range: range,
    },
    {
      label: "Speed",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "Speed",
      documentation: "Connection speed",
      detail: "number",
      range: range,
    },
    {
      label: "LastCheck",
      kind: monaco.languages.CompletionItemKind.Property,
      insertText: "LastCheck",
      documentation: "Last check timestamp",
      detail: "string",
      range: range,
    },
  ];

  return suggestions;
};
