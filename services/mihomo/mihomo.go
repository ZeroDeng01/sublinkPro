package mihomo

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sublink/node/protocol"
	"sublink/utils"
	"time"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
	"gopkg.in/yaml.v3"
)

// GetMihomoAdapter creates a Mihomo Proxy Adapter from a node link
func GetMihomoAdapter(nodeLink string) (constant.Proxy, error) {
	// 1. Parse node link to Proxy struct
	// We use a default OutputConfig as we only need the proxy connection info
	outputConfig := protocol.OutputConfig{
		Udp:  true,
		Cert: true, // Skip cert verify by default for better compatibility? Or false?
	}

	// Parse the link to get basic info
	// We need to construct a Urls struct
	_, err := url.Parse(nodeLink)
	if err != nil {
		return nil, fmt.Errorf("parse link error: %v", err)
	}

	// We need to handle the case where ParseNodeLink might be better, but LinkToProxy expects Urls struct
	// LinkToProxy handles various protocols
	proxyStruct, err := protocol.LinkToProxy(protocol.Urls{Url: nodeLink}, outputConfig)
	if err != nil {
		return nil, fmt.Errorf("convert link to proxy error: %v", err)
	}

	// 2. Convert Proxy struct to map[string]interface{} via YAML
	// This is because adapter.ParseProxy expects a map
	yamlBytes, err := yaml.Marshal(proxyStruct)
	if err != nil {
		return nil, fmt.Errorf("marshal proxy error: %v", err)
	}

	var proxyMap map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &proxyMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal proxy map error: %v", err)
	}

	// 3. Create Mihomo Proxy Adapter
	proxyAdapter, err := adapter.ParseProxy(proxyMap)
	if err != nil {
		return nil, fmt.Errorf("create mihomo adapter error: %v", err)
	}

	return proxyAdapter, nil
}

// MihomoDelayWithAdapter 使用 Mihomo 内置 URLTest 进行延迟测试
// 这是内部函数，直接调用 adapter 的 URLTest 方法
// includeHandshake: true 测量完整连接时间，false 使用 UnifiedDelay 模式排除握手
func MihomoDelayWithAdapter(proxyAdapter constant.Proxy, testUrl string, timeout time.Duration, includeHandshake bool) (latency int, err error) {
	// Recover from any panics and return error with zero latency
	defer func() {
		if r := recover(); r != nil {
			latency = 0
			err = fmt.Errorf("panic in MihomoDelayWithAdapter: %v", r)
		}
	}()

	if testUrl == "" {
		testUrl = "https://cp.cloudflare.com/generate_204"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 设置 UnifiedDelay 模式：
	// - includeHandshake=true -> UnifiedDelay=false（包含握手，单次请求）
	// - includeHandshake=false -> UnifiedDelay=true（排除握手，发两次请求取第二次）
	adapter.UnifiedDelay.Store(!includeHandshake)

	// 使用 Mihomo 内置的 URLTest 方法
	// expectedStatus 传 nil 表示接受任何成功状态码
	delay, err := proxyAdapter.URLTest(ctx, testUrl, nil)
	if err != nil {
		return 0, err
	}

	return int(delay), nil
}

// MihomoDelayTest 执行延迟测试，可选检测落地IP
// includeHandshake: true 测量完整连接时间，false 使用 UnifiedDelay 模式排除握手
// detectLandingIP: 是否检测落地IP
// landingIPUrl: IP查询服务URL，空则使用默认值 https://api.ipify.org
// 返回: latency(ms), landingIP(若未检测或失败则为空), error
func MihomoDelayTest(nodeLink string, testUrl string, timeout time.Duration, includeHandshake bool, detectLandingIP bool, landingIPUrl string) (latency int, landingIP string, err error) {
	// Recover from any panics
	defer func() {
		if r := recover(); r != nil {
			latency = 0
			landingIP = ""
			err = fmt.Errorf("panic in MihomoDelayTest: %v", r)
		}
	}()

	if testUrl == "" {
		testUrl = "http://cp.cloudflare.com/generate_204"
	}

	// 创建 adapter
	proxyAdapter, err := GetMihomoAdapter(nodeLink)
	if err != nil {
		return 0, "", err
	}

	// 执行延迟测试（使用 URLTest）
	latency, err = MihomoDelayWithAdapter(proxyAdapter, testUrl, timeout, includeHandshake)
	if err != nil {
		return 0, "", err
	}

	// 延迟测试成功后，如果需要检测落地IP
	if detectLandingIP {
		landingIP = fetchLandingIPWithAdapter(proxyAdapter, landingIPUrl)
	}

	return latency, landingIP, nil
}

// MihomoSpeedTest 执行速度测试，可选检测落地IP
// detectLandingIP: 是否检测落地IP
// landingIPUrl: IP查询服务URL，空则使用默认值 https://api.ipify.org
// speedRecordMode: 速度记录模式 "average"=平均速度, "peak"=峰值速度
// peakSampleInterval: 峰值采样间隔（毫秒），仅在peak模式下生效，范围50-200
// 返回: speed(MB/s), latency(ms), bytesDownloaded, landingIP(若未检测或失败则为空), error
func MihomoSpeedTest(nodeLink string, testUrl string, timeout time.Duration, detectLandingIP bool, landingIPUrl string, speedRecordMode string, peakSampleInterval int) (speed float64, latency int, bytesDownloaded int64, landingIP string, err error) {
	// Recover from any panics and return error with zero values
	defer func() {
		if r := recover(); r != nil {
			speed = 0
			latency = 0
			bytesDownloaded = 0
			landingIP = ""
			err = fmt.Errorf("panic in MihomoSpeedTest: %v", r)
		}
	}()

	// 默认值处理
	if speedRecordMode == "" {
		speedRecordMode = "average"
	}
	if peakSampleInterval < 50 {
		peakSampleInterval = 50
	} else if peakSampleInterval > 200 {
		peakSampleInterval = 200
	}

	proxyAdapter, err := GetMihomoAdapter(nodeLink)
	if err != nil {
		return 0, 0, 0, "", err
	}

	// 4. Perform Speed Test
	// We will try to download from testUrl
	if testUrl == "" {
		testUrl = "https://speed.cloudflare.com/__down?bytes=10000000" // Default 10MB
	}

	parsedUrl, err := url.Parse(testUrl)
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("parse test url error: %v", err)
	}

	portStr := parsedUrl.Port()
	if portStr == "" {
		if parsedUrl.Scheme == "https" {
			portStr = "443"
		} else {
			portStr = "80"
		}
	}

	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("invalid port: %v", err)
	}
	// Validate port range to prevent overflow
	if portInt < 0 || portInt > 65535 {
		return 0, 0, 0, "", fmt.Errorf("port out of range: %d", portInt)
	}
	port := uint16(portInt)

	metadata := &constant.Metadata{
		Host:    parsedUrl.Hostname(),
		DstPort: port,
		Type:    constant.HTTP,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	conn, err := proxyAdapter.DialContext(ctx, metadata)
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("dial error: %v", err)
	}
	// Close connection asynchronously to avoid blocking if it hangs
	defer func() {
		go func() {
			_ = conn.Close()
		}()
	}()

	// Calculate latency
	latency = int(time.Since(start).Milliseconds())

	// Create HTTP request
	req, err := http.NewRequest("GET", testUrl, nil)
	if err != nil {
		return 0, latency, 0, "", fmt.Errorf("create request error: %v", err)
	}
	req = req.WithContext(ctx)

	// We need to use the connection to send the request
	// Better approach: Use http.Client with a custom Transport that uses the proxy adapter.

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Recover from panics in DialContext
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic in DialContext: %v", r)
					}
				}()

				// Re-parse addr to get host and port for metadata
				h, pStr, splitErr := net.SplitHostPort(addr)
				if splitErr != nil {
					return nil, fmt.Errorf("split host port error: %v", splitErr)
				}

				pInt, atoiErr := strconv.Atoi(pStr)
				if atoiErr != nil {
					return nil, fmt.Errorf("invalid port string: %v", atoiErr)
				}

				// Validate port range
				if pInt < 0 || pInt > 65535 {
					return nil, fmt.Errorf("port out of range: %d", pInt)
				}
				p := uint16(pInt)

				md := &constant.Metadata{
					Host:    h,
					DstPort: p,
					Type:    constant.HTTP,
				}
				return proxyAdapter.DialContext(ctx, md)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: timeout,
	}

	resp, err := client.Get(testUrl)
	if err != nil {
		return 0, latency, 0, "", fmt.Errorf("http get error: %v", err)
	}
	defer resp.Body.Close()

	// Read body to measure speed
	// We can read up to N bytes or until EOF
	buf := make([]byte, 32*1024)
	var totalRead int64 // Changed to int64 to avoid overflow for large downloads
	readStart := time.Now()

	// 峰值速度采样相关变量
	var peakSpeed float64
	var lastSampleBytes int64
	var lastSampleTime time.Time
	var sampleTicker *time.Ticker
	var sampleDone chan struct{}

	if speedRecordMode == "peak" {
		lastSampleTime = readStart
		lastSampleBytes = 0
		sampleTicker = time.NewTicker(time.Duration(peakSampleInterval) * time.Millisecond)
		sampleDone = make(chan struct{})

		// 采样协程：按固定间隔计算瞬时速度
		go func() {
			defer sampleTicker.Stop()
			for {
				select {
				case <-sampleTicker.C:
					now := time.Now()
					currentBytes := totalRead
					elapsed := now.Sub(lastSampleTime).Seconds()
					if elapsed > 0 {
						// 计算瞬时速度 (MB/s)
						instantSpeed := float64(currentBytes-lastSampleBytes) / 1024 / 1024 / elapsed
						if instantSpeed > peakSpeed {
							peakSpeed = instantSpeed
						}
					}
					lastSampleBytes = currentBytes
					lastSampleTime = now
				case <-sampleDone:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	for {
		n, err := resp.Body.Read(buf)
		totalRead += int64(n)
		if err != nil {
			if err == io.EOF {
				break
			}
			// If context deadline exceeded (timeout), we consider it a successful test completion
			// because we want to measure speed over a fixed duration.
			if ctx.Err() == context.DeadlineExceeded || err == context.DeadlineExceeded || (err != nil && err.Error() == "context deadline exceeded") {
				break
			}
			// Check if it's a net.Error timeout
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			if sampleDone != nil {
				close(sampleDone)
			}
			return 0, latency, totalRead, "", fmt.Errorf("read body error: %v", err)
		}
		// Check timeout explicitly via context
		select {
		case <-ctx.Done():
			// Timeout reached, break loop to calculate speed
			goto CalculateSpeed
		default:
			// Continue reading
		}
	}

CalculateSpeed:

	// 停止采样协程
	if sampleDone != nil {
		close(sampleDone)
	}

	duration := time.Since(readStart)
	if duration.Seconds() == 0 {
		return 0, latency, totalRead, "", nil
	}

	// 最小有效下载量校验（10KB），避免因下载量过小导致速度虚高
	const minValidBytes int64 = 10 * 1024 // 10KB
	if totalRead < minValidBytes {
		return 0, latency, totalRead, "", fmt.Errorf("下载量过小 (%d 字节 < %d 字节)，结果不可靠", totalRead, minValidBytes)
	}

	// 根据模式选择返回值
	if speedRecordMode == "peak" && peakSpeed > 0 {
		// 使用峰值速度
		speed = peakSpeed
	} else {
		// 使用平均速度 (MB/s)
		speed = float64(totalRead) / 1024 / 1024 / duration.Seconds()
	}

	// 速度测试成功后，如果需要检测落地IP
	if detectLandingIP && speed > 0 {
		landingIP = fetchLandingIPWithAdapter(proxyAdapter, landingIPUrl)
	}

	return speed, latency, totalRead, landingIP, nil
}

// fetchLandingIPWithAdapter 使用已有adapter获取落地IP（内部辅助函数）
// 固定1秒超时，失败静默返回空字符串不影响主流程
func fetchLandingIPWithAdapter(proxyAdapter constant.Proxy, ipUrl string) string {
	// Recover from any panics
	defer func() {
		if r := recover(); r != nil {
			// 静默处理，不影响主流程
		}
	}()

	// 默认IP查询接口
	if ipUrl == "" {
		ipUrl = "https://api.ipify.org"
	}

	// 固定3秒超时（慢速节点需要更长时间）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 复用proxyAdapter创建HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(dialCtx context.Context, network, addr string) (net.Conn, error) {
				h, pStr, splitErr := net.SplitHostPort(addr)
				if splitErr != nil {
					return nil, splitErr
				}

				pInt, atoiErr := strconv.Atoi(pStr)
				if atoiErr != nil {
					return nil, atoiErr
				}

				if pInt < 0 || pInt > 65535 {
					return nil, fmt.Errorf("port out of range: %d", pInt)
				}

				md := &constant.Metadata{
					Host:    h,
					DstPort: uint16(pInt),
					Type:    constant.HTTP,
				}
				return proxyAdapter.DialContext(dialCtx, md)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", ipUrl, nil)
	if err != nil {
		utils.Error("落地IP检测: 创建请求失败: %v", err)
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		utils.Error("落地IP检测: 请求失败: %v (URL: %s)", err, ipUrl)
		return ""
	}
	defer resp.Body.Close()

	// 限制读取最多64字节（IP地址不会超过这个长度）
	body := make([]byte, 64)
	n, _ := resp.Body.Read(body)

	return strings.TrimSpace(string(body[:n]))
}
