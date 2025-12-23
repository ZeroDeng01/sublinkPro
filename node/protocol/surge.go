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
	"sublink/utils"
)

func EncodeSurge(urls []string, config OutputConfig) (string, error) {
	var proxys, groups []string

	// 辅助函数：根据 HostMap 替换服务器地址
	replaceHost := func(server string) string {
		if config.ReplaceServerWithHost && len(config.HostMap) > 0 {
			if ip, exists := config.HostMap[server]; exists {
				return ip
			}
		}
		return server
	}

	for _, link := range urls {
		Scheme := strings.Split(link, "://")[0]
		switch {
		case Scheme == "ss":
			ss, err := DecodeSSURL(link)
			if err != nil {
				log.Println(err)
				continue
			}
			server := replaceHost(ss.Server)
			proxy := map[string]interface{}{
				"name":     ss.Name,
				"server":   server,
				"port":     utils.GetPortInt(ss.Port),
				"cipher":   ss.Param.Cipher,
				"password": ss.Param.Password,
				"udp":      config.Udp,
			}
			ssproxy := fmt.Sprintf("%s = ss, %s, %d, encrypt-method=%s, password=%s, udp-relay=%t",
				proxy["name"], proxy["server"], proxy["port"], proxy["cipher"], proxy["password"], proxy["udp"])
			groups = append(groups, ss.Name)
			proxys = append(proxys, ssproxy)
		case Scheme == "vmess":
			vmess, err := DecodeVMESSURL(link)
			if err != nil {
				log.Println(err)
				continue
			}
			tls := false
			if vmess.Tls != "none" && vmess.Tls != "" {
				tls = true
			}
			port, _ := convertToInt(vmess.Port)
			server := replaceHost(vmess.Add)
			proxy := map[string]interface{}{
				"name":             vmess.Ps,
				"server":           server,
				"port":             port,
				"uuid":             vmess.Id,
				"tls":              tls,
				"network":          vmess.Net,
				"ws-path":          vmess.Path,
				"ws-host":          vmess.Host,
				"udp":              config.Udp,
				"skip-cert-verify": config.Cert,
			}
			vmessproxy := fmt.Sprintf("%s = vmess, %s, %d, username=%s , tls=%t, vmess-aead=true,  udp-relay=%t , skip-cert-verify=%t",
				proxy["name"], proxy["server"], proxy["port"], proxy["uuid"], proxy["tls"], proxy["udp"], proxy["skip-cert-verify"])
			if vmess.Net == "ws" {
				vmessproxy = fmt.Sprintf("%s, ws=true,ws-path=%s", vmessproxy, proxy["ws-path"])
				if vmess.Host != "" && vmess.Host != "none" {
					vmessproxy = fmt.Sprintf("%s, ws-headers=Host:%s", vmessproxy, proxy["ws-host"])
				}
			}
			if vmess.Sni != "" {
				vmessproxy = fmt.Sprintf("%s, sni=%s", vmessproxy, vmess.Sni)
			}
			groups = append(groups, vmess.Ps)
			proxys = append(proxys, vmessproxy)
		case Scheme == "trojan":
			trojan, err := DecodeTrojanURL(link)
			if err != nil {
				log.Println(err)
				continue
			}
			server := replaceHost(trojan.Hostname)
			proxy := map[string]interface{}{
				"name":             trojan.Name,
				"server":           server,
				"port":             utils.GetPortInt(trojan.Port),
				"password":         trojan.Password,
				"udp":              config.Udp,
				"skip-cert-verify": config.Cert,
			}
			trojanproxy := fmt.Sprintf("%s = trojan, %s, %d, password=%s, udp-relay=%t, skip-cert-verify=%t",
				proxy["name"], proxy["server"], proxy["port"], proxy["password"], proxy["udp"], proxy["skip-cert-verify"])
			if trojan.Query.Sni != "" {
				trojanproxy = fmt.Sprintf("%s, sni=%s", trojanproxy, trojan.Query.Sni)

			}
			groups = append(groups, trojan.Name)
			proxys = append(proxys, trojanproxy)
		case Scheme == "hysteria2" || Scheme == "hy2":
			hy2, err := DecodeHY2URL(link)
			if err != nil {
				log.Println(err)
				continue
			}
			server := replaceHost(hy2.Host)
			proxy := map[string]interface{}{
				"name":             hy2.Name,
				"server":           server,
				"port":             utils.GetPortInt(hy2.Port),
				"password":         hy2.Password,
				"udp":              config.Udp,
				"skip-cert-verify": config.Cert,
			}
			hy2proxy := fmt.Sprintf("%s = hysteria2, %s, %d, password=%s, udp-relay=%t, skip-cert-verify=%t",
				proxy["name"], proxy["server"], proxy["port"], proxy["password"], proxy["udp"], proxy["skip-cert-verify"])
			if hy2.Sni != "" {
				hy2proxy = fmt.Sprintf("%s, sni=%s", hy2proxy, hy2.Sni)

			}
			groups = append(groups, hy2.Name)
			proxys = append(proxys, hy2proxy)
		case Scheme == "tuic":
			tuic, err := DecodeTuicURL(link)
			if err != nil {
				log.Println(err)
				continue
			}
			server := replaceHost(tuic.Host)
			proxy := map[string]interface{}{
				"name":             tuic.Name,
				"server":           server,
				"port":             utils.GetPortInt(tuic.Port),
				"password":         tuic.Password,
				"udp":              config.Udp,
				"skip-cert-verify": config.Cert,
			}
			tuicproxy := fmt.Sprintf("%s = tuic, %s, %d, token=%s, udp-relay=%t, skip-cert-verify=%t",
				proxy["name"], proxy["server"], proxy["port"], proxy["password"], proxy["udp"], proxy["skip-cert-verify"])
			groups = append(groups, tuic.Name)
			proxys = append(proxys, tuicproxy)
		}
	}
	return DecodeSurge(proxys, groups, config.Surge)
}
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

			// 没有自动匹配参数，追加所有节点
			line = strings.TrimSpace(line) + ", " + grouplist
			// 确保代理组有有效节点
			line = ensureProxyGroupHasProxies(line)
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n"), nil
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
