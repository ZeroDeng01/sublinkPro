<div align="center">
<img src="webs/src/assets/logo.png" width="150px" height="150px" />
</div>

<div align="center">
  <img src="https://img.shields.io/badge/Vue-5.0.8-brightgreen.svg"/>
  <img src="https://img.shields.io/badge/Go-1.24.3-green.svg"/>
  <img src="https://img.shields.io/badge/Element%20Plus-2.6.1-blue.svg"/>
  <img src="https://img.shields.io/badge/license-MIT-green.svg"/>
  <div align="center"> <a href="README.md">中文</a> | English</div>


</div>

# Project Introduction

`sublinkE` is an open-source node subscription conversion and generation management system based on secondary development of  [sublinkX](https://github.com/gooaclok819/sublinkX) /[sublinkE](https://github.com/eun1e/sublinkE)  .It is recommended that users give priority to referring to and using the original project, and we sincerely thank the original author for their efforts and contributions.

- Frontend based on [vue3-element-admin](https://github.com/youlaitech/vue3-element-admin);
- Backend using Go + Gin + Gorm;
- Default account: admin Password: 123456, please make sure to change it after installation;

# Modifications


- [x] Fix some page bugs
- [x] Support Clash `dialer-proxy` attribute
- [x] Allow adding and using API KEY to access API
- [x] Import and schedule updates of nodes from subscription links
- [x] Support AnyTLS and Socks5 protocols
- [x] Subscription node sorting
- [x] Support plugin extensions (experimental)
- [x] IP Blacklist/Whitelist Subscription Support
- [ ] ...

# Project Features

- High flexibility and security, supporting subscription record access and simple configuration management;
- Support for multiple client protocols and formats, including:
    - v2ray (base64 universal format)
    - clash (support ss, ssr, trojan, vmess, vless, hy, hy2, tuic, AnyTLS, Socks5)
    - surge (support ss, trojan, vmess, hy2, tuic)
- Added token authorization and subscription import functionality, enhancing security and convenience.

# 📦 Installation Guide

## 🚀 Run with Docker
```bash
docker run --name sublinke -p 8000:8000 \
-v $PWD/db:/app/db \
-v $PWD/template:/app/template \
-v $PWD/logs:/app/logs \
-v $PWD/plugins:/app/plugins \
-d zerodeng/sublink-pro 
```

## ⚡ One-click Installation
```bash
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh   && sh install.sh
```

> ⚠ **Note**  
> When running the one-click installation script on **Alpine Linux**, the plugin module may not work properly because Alpine uses `musl` instead of `glibc`.  
> It is recommended to use **Docker deployment** for the best compatibility, or choose distributions such as **Debian / Ubuntu**.


# Plugin Description

`sublinkPro` provides a flexible plugin system, allowing developers to extend system functionality without modifying core code.

## Plugin Development Guide

### Basic Steps

1. **Create Plugin File**: Refer to `plugins_examples/email_plugin.go` to write custom plugins
2. **Compile Plugin**: Use `plugins_examples/build_plugin.sh email_plugin.go` to compile into a `.so` file
3. **Deploy Plugin**: Place the compiled `.so` file into the `plugins` directory

### Plugin Interface Implementation

All plugins must implement the `plugins.Plugin` interface, which includes the following core methods:

```go
// Required methods
Name() string                           // Plugin name
Version() string                        // Plugin version
Description() string                    // Plugin description
DefaultConfig() map[string]interface{}  // Default configuration
SetConfig(map[string]interface{})       // Set configuration
Init() error                            // Initialize
Close() error                           // Cleanup

// Event handling method (API event listening)
OnAPIEvent(ctx *gin.Context, event plugins.EventType, path string, 
           statusCode int, requestBody interface{}, 
           responseBody interface{}) error

// Declare API paths and event types the plugin is interested in
InterestedAPIs() []string
InterestedEvents() []plugins.EventType
```

### Plugin Examples

The system includes the following example plugins for reference(May become invalid after version updates; it's recommended to compile it yourself.):

| Plugin Name | Description | Source Code | Compiled Version |
|-------------|-------------|------------|-----------------|
| **Email Notification Plugin** | Monitors login events and sends email notifications | [email_plugin.go](https://github.com/ZeroDeng01/sublinkPro/blob/main/plugins_examples/email_plugin.go) | [Download .so file](https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/main/plugins_examples/email_plugin.so) |

### Plugin Configuration and Management

Plugins can be managed through the Web interface:
- Enable/disable plugins
- Configure plugin parameters
- View plugin status

## Developing Custom Plugins

Custom plugin development workflow:

1. Create a Go plugin file, implementing the `plugins.Plugin` interface
2. Export the `GetPlugin()` function, returning the plugin instance
3. Define API paths and event types the plugin is concerned with
4. Implement event handling logic
5. Use the build script to compile the plugin

```bash
# Compile plugin
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/main/plugins_examples/build_plugin.sh
chmod +x build_plugin.sh
./build_plugin.sh your_plugin.go
# Copy the generated .so file to the plugins directory
cp your_plugin.so ../plugins/
```

For more advanced features and detailed API documentation, please refer to the code examples.

# Project Preview

![Preview1](webs/src/assets/1.png)
![Preview2](webs/src/assets/2.png)
![Preview3](webs/src/assets/3.png)
![Preview4](webs/src/assets/4.png)
![Preview5](webs/src/assets/5.png)
![Preview6](webs/src/assets/6.png)
