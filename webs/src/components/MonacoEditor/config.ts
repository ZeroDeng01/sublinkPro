import * as monaco from "monaco-editor";

// Global flag to ensure configuration only runs once
let isConfigured = false;

/**
 * Configure Monaco Editor global settings
 * This should be called before any editor instance is created
 */
export const configureMonacoEditor = () => {
  if (isConfigured) {
    return;
  }
  isConfigured = true;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const javascriptDefaults = (monaco.languages.typescript as any)
    .javascriptDefaults;

  // Enable diagnostics for syntax checking
  javascriptDefaults.setDiagnosticsOptions({
    noSemanticValidation: false, // Enable semantic validation
    noSyntaxValidation: false, // Enable syntax validation
    diagnosticCodesToIgnore: [1108], // Ignore "return statement can only be used within a function body"
  });

  // Configure compiler options
  javascriptDefaults.setCompilerOptions({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    target: (monaco.languages.typescript as any).ScriptTarget.ES2020,
    allowNonTsExtensions: true,
    allowJs: true,
    checkJs: true, // Enable type checking for JavaScript
  });

  // Enable eager model sync for better IntelliSense
  javascriptDefaults.setEagerModelSync(true);

  // Inject custom type definitions
  javascriptDefaults.addExtraLib(
    `
    /**
     * Represents a node in the subscription system.
     */
    interface Node {
      /** Unique identifier for the node */
      ID: number;
      /** Full link string */
      Link: string;
      /** Display name of the node */
      Name: string;
      /** Link identifier name */
      LinkName: string;
      /** Link address */
      LinkAddress: string;
      /** Link host */
      LinkHost: string;
      /** Link port */
      LinkPort: string;
      /** Dialer proxy name */
      DialerProxyName: string;
      /** Creation date */
      CreateDate: string;
      /** Source identifier */
      Source: string;
      /** Source ID number */
      SourceID: number;
      /** Node group */
      Group: string;
      /** Connection speed */
      Speed: number;
      /** Last check timestamp */
      LastCheck: string;
    }

    /**
     * Filter nodes based on custom logic.
     * This is the main entry point for node filtering scripts.
     * @param nodes - Array of nodes to filter
     * @param clientType - The type of client (e.g., "v2ray", "clash", "surge")
     * @returns Filtered array of nodes
     * @example
     * function filterNode(nodes, clientType) {
     *   return nodes.filter(node => node.Speed > 100);
     * }
     */
    declare function filterNode(nodes: Node[], clientType: string): Node[];

    /**
     * Modify the subscription content.
     * This is the main entry point for subscription modification scripts.
     * @param input - The original subscription content as a string
     * @param clientType - The type of client (e.g., "v2ray", "clash", "surge")
     * @returns Modified subscription content
     * @example
     * function subMod(input, clientType) {
     *   return input.replace(/old/g, 'new');
     * }
     */
    declare function subMod(input: string, clientType: string): string;

    /**
     * Console object for logging. Available in the script execution environment.
     */
    declare const console: {
      /** Log a message */
      log(...args: any[]): void;
      /** Log an info message */
      info(...args: any[]): void;
      /** Log a warning message */
      warn(...args: any[]): void;
      /** Log an error message */
      error(...args: any[]): void;
    };
    `,
    "ts:filename/custom-types.d.ts"
  );
};

// Auto-configure when module is imported
configureMonacoEditor();
