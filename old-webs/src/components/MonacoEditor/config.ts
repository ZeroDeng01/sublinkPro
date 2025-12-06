import * as monaco from "monaco-editor";

// 全局标志，确保配置只运行一次
let isConfigured = false;

/**
 * 配置 Monaco Editor 全局设置
 * 应在创建任何编辑器实例之前调用
 */
export const configureMonacoEditor = () => {
  if (isConfigured) {
    return;
  }
  isConfigured = true;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const javascriptDefaults = (monaco.languages.typescript as any)
    .javascriptDefaults;

  // 启用语法检查诊断
  javascriptDefaults.setDiagnosticsOptions({
    noSemanticValidation: false, // 启用语义验证
    noSyntaxValidation: false, // 启用语法验证
    diagnosticCodesToIgnore: [1108], // 忽略 "return 语句只能在函数体内使用"
  });

  // 配置编译器选项
  javascriptDefaults.setCompilerOptions({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    target: (monaco.languages.typescript as any).ScriptTarget.ES2020,
    allowNonTsExtensions: true,
    allowJs: true,
    checkJs: true, // 为 JavaScript 启用类型检查
  });

  // 启用即时模型同步以获得更好的智能提示
  javascriptDefaults.setEagerModelSync(true);

  // 注入自定义类型定义
  javascriptDefaults.addExtraLib(
    `
    /**
     * 表示订阅系统中的节点
     */
    interface Node {
      /** 节点的唯一标识符 */
      ID: number;
      /** 完整链接字符串 */
      Link: string;
      /** 节点显示名称 */
      Name: string;
      /** 链接标识名称 */
      LinkName: string;
      /** 链接地址 */
      LinkAddress: string;
      /** 链接主机 */
      LinkHost: string;
      /** 链接端口 */
      LinkPort: string;
      /** 拨号代理名称 */
      DialerProxyName: string;
      /** 创建日期 */
      CreateDate: string;
      /** 来源标识符 */
      Source: string;
      /** 来源 ID 编号 */
      SourceID: number;
      /** 节点分组 */
      Group: string;
      /** 节点速度MB/s */
      Speed: number;
      /** 节点延迟ms */
      DelayTime: number;
      /** 最后检查时间戳 */
      LastCheck: string;
    }

    /**
     * 基于自定义逻辑过滤节点
     * 这是节点过滤脚本的主入口函数
     * @param nodes - 要过滤的节点数组
     * @param clientType - 客户端类型（例如："v2ray"、"clash"、"surge"）
     * @returns 过滤后的节点数组
     * @example
     * function filterNode(nodes, clientType) {
     *   return nodes.filter(node => node.Speed > 100);
     * }
     */
    declare function filterNode(nodes: Node[], clientType: string): Node[];

    /**
     * 修改订阅内容
     * 这是订阅修改脚本的主入口函数
     * @param input - 原始订阅内容字符串
     * @param clientType - 客户端类型（例如："v2ray"、"clash"、"surge"）
     * @returns 修改后的订阅内容
     * @example
     * function subMod(input, clientType) {
     *   return input.replace(/old/g, 'new');
     * }
     */
    declare function subMod(input: string, clientType: string): string;

    /**
     * 用于日志记录的 Console 对象。在脚本执行环境中可用。
     */
    declare const console: {
      /** 记录一条消息 */
      log(...args: any[]): void;
      /** 记录一条信息消息 */
      info(...args: any[]): void;
      /** 记录一条警告消息 */
      warn(...args: any[]): void;
      /** 记录一条错误消息 */
      error(...args: any[]): void;
    };
    `,
    "ts:filename/custom-types.d.ts"
  );
};

// 模块导入时自动配置
configureMonacoEditor();
