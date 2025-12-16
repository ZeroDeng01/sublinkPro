package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/metacubex/mihomo/constant"
)

// GetMihomoAdapterFunc 获取 mihomo 适配器的函数类型
// 需要在 main.go 或 init 阶段设置
var GetMihomoAdapterFunc func(nodeLink string) (constant.Proxy, error)

// GetBestProxyNodeFunc 获取最佳代理节点的函数类型
// 需要在 main.go 或 init 阶段设置
var GetBestProxyNodeFunc func() (link string, name string, err error)

// CreateProxyHTTPClient 创建带代理的HTTP客户端
// useProxy: 是否使用代理
// proxyLink: 代理节点链接，为空时自动选择最佳代理
// timeout: 请求超时时间
// 返回: HTTP客户端, 使用的代理节点链接, 错误
func CreateProxyHTTPClient(useProxy bool, proxyLink string, timeout time.Duration) (*http.Client, string, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	if !useProxy {
		return client, "", nil
	}

	var proxyNodeLink string

	if proxyLink != "" {
		// 使用指定的代理链接
		proxyNodeLink = proxyLink
		Info("使用指定代理下载")
	} else if GetBestProxyNodeFunc != nil {
		// 如果没有指定代理，尝试自动选择最佳代理
		link, name, err := GetBestProxyNodeFunc()
		if err == nil && link != "" {
			Info("自动选择最佳代理节点: %s", name)
			proxyNodeLink = link
		}
	}

	if proxyNodeLink == "" {
		Warn("未找到可用代理，将直接下载")
		return client, "", nil
	}

	if GetMihomoAdapterFunc == nil {
		Warn("Mihomo 适配器未初始化，将直接下载")
		return client, "", nil
	}

	// 使用 mihomo 内核创建代理适配器
	proxyAdapter, err := GetMihomoAdapterFunc(proxyNodeLink)
	if err != nil {
		Error("创建 mihomo 代理适配器失败: %v，将直接下载", err)
		return client, "", nil
	}

	Info("使用 mihomo 内核代理下载")
	// 创建自定义 Transport，使用 mihomo adapter 进行代理连接
	client.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 解析地址获取主机和端口
			host, portStr, splitErr := net.SplitHostPort(addr)
			if splitErr != nil {
				return nil, fmt.Errorf("split host port error: %v", splitErr)
			}

			portInt, atoiErr := strconv.Atoi(portStr)
			if atoiErr != nil {
				return nil, fmt.Errorf("invalid port: %v", atoiErr)
			}

			// 验证端口范围
			if portInt < 0 || portInt > 65535 {
				return nil, fmt.Errorf("port out of range: %d", portInt)
			}

			// 创建 mihomo metadata
			metadata := &constant.Metadata{
				Host:    host,
				DstPort: uint16(portInt),
				Type:    constant.HTTP,
			}

			// 使用 mihomo adapter 建立连接
			return proxyAdapter.DialContext(ctx, metadata)
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return client, proxyNodeLink, nil
}

// FetchWithProxy 通过代理获取远程内容
// url: 目标URL
// useProxy: 是否使用代理
// proxyLink: 代理节点链接（可选，为空自动选择）
// timeout: 超时时间
// userAgent: 请求的 User-Agent (可选)
// 返回: 响应内容, 错误
func FetchWithProxy(url string, useProxy bool, proxyLink string, timeout time.Duration, userAgent string) ([]byte, error) {
	client, _, err := CreateProxyHTTPClient(useProxy, proxyLink, timeout)
	if err != nil {
		return nil, err
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %v", err)
	}

	// 设置 User-Agent
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", "Clash")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %v", err)
	}

	return body, nil
}

// FetchStringWithProxy 通过代理获取远程内容（返回字符串）
// url: 目标URL
// useProxy: 是否使用代理
// proxyLink: 代理节点链接（可选，为空自动选择）
// timeout: 超时时间
// 返回: 响应内容字符串, 错误
func FetchStringWithProxy(url string, useProxy bool, proxyLink string, timeout time.Duration) (string, error) {
	data, err := FetchWithProxy(url, useProxy, proxyLink, timeout, "")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
