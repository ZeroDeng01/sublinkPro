package protocol

// HTTP/HTTPS协议使用示例

// 示例1: HTTP代理链接
// http://username:password@server:port#节点名称
// 例如: http://user:pass@192.168.1.1:8080#我的HTTP代理

// 示例2: HTTPS代理链接（带TLS）
// https://username:password@server:port#节点名称
// 例如: https://user:pass@192.168.1.1:443#我的HTTPS代理

// 示例3: HTTPS代理链接（带跳过证书验证和SNI）
// https://username:password@server:port?skip-cert-verify=true&sni=example.com#节点名称
// 例如: https://user:pass@192.168.1.1:8443?skip-cert-verify=true&sni=example.com#我的HTTPS代理

// 示例4: HTTP代理链接（无认证）
// http://server:port#节点名称
// 例如: http://192.168.1.1:8080#公开HTTP代理

// 示例5: HTTPS代理链接（无认证）
// https://server:port#节点名称
// 例如: https://192.168.1.1:443#公开HTTPS代理

// Clash配置示例:
// - name: "🇨🇳 直连"
//   type: http
//   server: xxx.xxx.cn
//   port: 8860
//   username: [zhanghao]
//   password: [mima]
//   tls: true
//   skip-cert-verify: false
//   sni: xxx.xxx.cn

// 支持的URL参数:
// - skip-cert-verify: 是否跳过证书验证 (true/false)
// - sni: 服务器名称指示 (SNI)

// 默认端口:
// - HTTP: 80
// - HTTPS: 443