package protocol

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sublink/cache"
)

// appendSurgeSSPlugin 将 SS 插件配置转换为 Surge 格式并追加到代理字符串
// Surge 主要支持 obfs (simple-obfs) 插件
func appendSurgeSSPlugin(proxy string, plugin SsPlugin) string {
	if plugin.Name == "" {
		return proxy
	}

	switch plugin.Name {
	case "obfs", "obfs-local", "simple-obfs":
		// Surge 格式: obfs=http/tls, obfs-host=xxx
		mode := plugin.Mode
		if mode == "" {
			mode = "http"
		}
		proxy = fmt.Sprintf("%s, obfs=%s", proxy, mode)
		if plugin.Host != "" {
			proxy = fmt.Sprintf("%s, obfs-host=%s", proxy, plugin.Host)
		}

	case "v2ray-plugin":
		// v2ray-plugin 在 Surge 中作为 ws 传输处理
		if plugin.Mode == "websocket" {
			proxy = fmt.Sprintf("%s, ws=true", proxy)
			if plugin.Path != "" {
				proxy = fmt.Sprintf("%s, ws-path=%s", proxy, plugin.Path)
			}
			if plugin.Host != "" {
				proxy = fmt.Sprintf("%s, ws-headers=Host:%s", proxy, plugin.Host)
			}
			if plugin.Tls {
				proxy = fmt.Sprintf("%s, tls=true", proxy)
			}
		}

		// shadow-tls、restls、kcptun 等插件 Surge 不原生支持，跳过
	}

	return proxy
}

// EncodeSurge 将节点链接批量转换为 Surge 节点定义，并交给模板合并逻辑统一落盘。
func EncodeSurge(urls []string, config OutputConfig) (string, error) {
	var proxys, groups []string

	for _, link := range urls {
		protocol := detectProtocol(link)
		if protocol == nil {
			continue
		}
		surgeCapable, ok := protocol.(SurgeCapable)
		if !ok {
			continue
		}
		proxyLine, groupName, err := surgeCapable.ToSurgeLine(link, config)
		if err != nil {
			log.Println(err)
			continue
		}
		if proxyLine != "" {
			proxys = append(proxys, proxyLine)
		}
		if groupName != "" {
			groups = append(groups, groupName)
		}
	}
	return DecodeSurge(proxys, groups, config.Surge)
}

func replaceSurgeHost(server string, config OutputConfig) string {
	if config.ReplaceServerWithHost && len(config.HostMap) > 0 {
		if ip, exists := config.HostMap[server]; exists {
			return ip
		}
	}
	return server
}

// DecodeSurge 读取 Surge 模板并合并节点与代理组。
// 该流程会尽量保留模板中的自动匹配组语义，只在需要时补节点或 DIRECT 后备项。
func DecodeSurge(proxys, groups []string, file string) (string, error) {
	var surge []byte
	var err error
	if strings.Contains(file, "://") {
		resp, err := http.Get(file)
		if err != nil {
			log.Println("http.Get error", err)
			return "", err
		}
		defer resp.Body.Close()
		surge, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error: %v", err)
			return "", err
		}
	} else {
		// 优先从缓存读取模板内容（本地文件使用缓存）
		filename := filepath.Base(file)
		if cached, ok := cache.GetTemplateContent(filename); ok {
			surge = []byte(cached)
		} else {
			surge, err = os.ReadFile(file)
			if err != nil {
				log.Println(err)
				return "", err
			}
			// 写入缓存
			cache.SetTemplateContent(filename, string(surge))
		}
	}

	// 按行处理模板文件
	lines := strings.Split(string(surge), "\n")
	var result []string
	currentSection := ""
	grouplist := strings.Join(groups, ", ")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 检测 section 标记
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			currentSection = trimmedLine
			result = append(result, line)

			// 在 [Proxy] section 后立即插入所有节点
			if currentSection == "[Proxy]" {
				for _, proxy := range proxys {
					result = append(result, proxy)
				}
			}
			continue
		}

		// 处理 [Proxy Group] section 中的代理组行
		if currentSection == "[Proxy Group]" && strings.Contains(line, "=") && trimmedLine != "" {
			// 如果已有 include-all-proxies，说明使用自动节点匹配模式，跳过节点插入
			// policy-regex-filter 需要 include-all-proxies 为前提
			// 这样可以减小配置文件大小，让客户端自动包含/过滤节点
			if strings.Contains(line, "include-all-proxies") {
				result = append(result, line)
				continue
			}

			// 关键逻辑：只对没有现有代理的组追加节点
			// 如果已有代理列表（组引用如 🚀 节点选择、DIRECT 等），保持不变
			// 这符合 ACL4SSR 的设计：只有使用 .* 的组才需要包含所有节点
			hasExistingProxies := surgeGroupHasProxies(line)
			if !hasExistingProxies {
				// 没有任何代理，追加所有节点
				line = strings.TrimSpace(line) + ", " + grouplist
				// 确保代理组有有效节点
				line = ensureProxyGroupHasProxies(line)
			}
			// 已有代理的组保持不变
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n"), nil
}

// surgeGroupHasProxies 检查 Surge 代理组行是否已有代理
// 格式: GroupName = type, proxy1, proxy2, ... 或 GroupName = type, url=xxx, ...
// 返回 true 如果已有代理（不包括 url= 等参数）
func surgeGroupHasProxies(line string) bool {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return false
	}
	afterEquals := strings.TrimSpace(parts[1])

	// 找到类型后的第一个逗号
	commaIndex := strings.Index(afterEquals, ",")
	if commaIndex == -1 {
		// 只有类型，没有任何代理
		return false
	}

	// 检查逗号后面的内容
	afterType := strings.TrimSpace(afterEquals[commaIndex+1:])
	if afterType == "" {
		return false
	}

	// 检查是否只有参数（url=, interval=, policy-regex-filter= 等）而没有代理
	// 这些参数通常以 xxx= 开头
	parts2 := strings.Split(afterType, ",")
	for _, part := range parts2 {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		// 如果不是参数格式（xxx=yyy），则认为是代理名称
		if !strings.Contains(trimmed, "=") {
			return true
		}
	}

	return false
}

// ensureProxyGroupHasProxies 检查 Surge 代理组行是否有有效节点
// 如果没有有效节点，追加 DIRECT 作为后备
// 格式: GroupName = type, proxy1, proxy2, ...
func ensureProxyGroupHasProxies(line string) string {
	// 分割行，检查 = 后面的内容
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return line
	}
	afterEquals := strings.TrimSpace(parts[1])

	// 找到类型后的第一个逗号
	commaIndex := strings.Index(afterEquals, ",")
	if commaIndex == -1 {
		// 只有类型，没有任何代理
		return line + ", DIRECT"
	}

	// 检查逗号后是否有有效内容
	afterType := strings.TrimSpace(afterEquals[commaIndex+1:])

	// 处理末尾多余的逗号和空格
	afterType = strings.TrimRight(afterType, ", ")

	if afterType == "" {
		// 清理末尾的逗号和空格，然后追加 DIRECT
		cleanLine := strings.TrimRight(line, ", ")
		return cleanLine + ", DIRECT"
	}

	return line
}
