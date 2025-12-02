package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sublink/node"
	"sublink/utils"
	"time"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
	"gopkg.in/yaml.v3"
)

// MihomoSpeedTest performs a true speed test using Mihomo adapter
// Returns speed in MB/s and latency in ms
func MihomoSpeedTest(nodeLink string, testUrl string, timeout time.Duration) (float64, int, error) {
	// 1. Parse node link to Proxy struct
	// We use a default SqlConfig as we only need the proxy connection info
	sqlConfig := utils.SqlConfig{
		Udp:  true,
		Cert: true, // Skip cert verify by default for better compatibility? Or false?
	}

	// Parse the link to get basic info
	// We need to construct a Urls struct
	_, err := url.Parse(nodeLink)
	if err != nil {
		return 0, 0, fmt.Errorf("parse link error: %v", err)
	}

	// We need to handle the case where ParseNodeLink might be better, but LinkToProxy expects Urls struct
	// LinkToProxy handles various protocols
	proxyStruct, err := node.LinkToProxy(node.Urls{Url: nodeLink}, sqlConfig)
	if err != nil {
		return 0, 0, fmt.Errorf("convert link to proxy error: %v", err)
	}

	// 2. Convert Proxy struct to map[string]interface{} via YAML
	// This is because adapter.ParseProxy expects a map
	yamlBytes, err := yaml.Marshal(proxyStruct)
	if err != nil {
		return 0, 0, fmt.Errorf("marshal proxy error: %v", err)
	}

	var proxyMap map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &proxyMap)
	if err != nil {
		return 0, 0, fmt.Errorf("unmarshal proxy map error: %v", err)
	}

	// 3. Create Mihomo Proxy Adapter
	proxyAdapter, err := adapter.ParseProxy(proxyMap)
	if err != nil {
		return 0, 0, fmt.Errorf("create mihomo adapter error: %v", err)
	}

	// 4. Perform Speed Test
	// We will try to download from testUrl
	if testUrl == "" {
		testUrl = "https://speed.cloudflare.com/__down?bytes=10000000" // Default 10MB
	}

	parsedUrl, err := url.Parse(testUrl)
	if err != nil {
		return 0, 0, fmt.Errorf("parse test url error: %v", err)
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
		return 0, 0, fmt.Errorf("invalid port: %v", err)
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
		return 0, 0, fmt.Errorf("dial error: %v", err)
	}
	defer conn.Close()

	// Calculate latency
	latency := time.Since(start).Milliseconds()

	// Create HTTP request
	req, err := http.NewRequest("GET", testUrl, nil)
	if err != nil {
		return 0, int(latency), fmt.Errorf("create request error: %v", err)
	}
	req = req.WithContext(ctx)

	// We need to use the connection to send the request
	// Better approach: Use http.Client with a custom Transport that uses the proxy adapter.

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Ignore addr, use the proxy adapter to dial the target
				// But wait, the addr passed here is the target address.
				// proxyAdapter.DialContext needs metadata.

				// Re-parse addr to get host and port for metadata
				h, pStr, _ := net.SplitHostPort(addr)
				pInt, _ := strconv.Atoi(pStr)
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
		return 0, int(latency), fmt.Errorf("http get error: %v", err)
	}
	defer resp.Body.Close()

	// Read body to measure speed
	// We can read up to N bytes or until EOF
	buf := make([]byte, 32*1024)
	totalRead := 0
	readStart := time.Now()

	for {
		n, err := resp.Body.Read(buf)
		totalRead += n
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
			return 0, int(latency), fmt.Errorf("read body error: %v", err)
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

	duration := time.Since(readStart)
	if duration.Seconds() == 0 {
		return 0, int(latency), nil
	}

	// Speed in MB/s
	speed := float64(totalRead) / 1024 / 1024 / duration.Seconds()

	return speed, int(latency), nil
}
