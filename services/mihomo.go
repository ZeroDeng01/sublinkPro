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

// GetMihomoAdapter creates a Mihomo Proxy Adapter from a node link
func GetMihomoAdapter(nodeLink string) (constant.Proxy, error) {
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
		return nil, fmt.Errorf("parse link error: %v", err)
	}

	// We need to handle the case where ParseNodeLink might be better, but LinkToProxy expects Urls struct
	// LinkToProxy handles various protocols
	proxyStruct, err := node.LinkToProxy(node.Urls{Url: nodeLink}, sqlConfig)
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

// MihomoDelay performs a latency test using Mihomo adapter (Protocol-aware)
// Returns latency in ms
func MihomoDelay(nodeLink string, testUrl string, timeout time.Duration) (latency int, err error) {
	// Recover from any panics and return error with zero latency
	defer func() {
		if r := recover(); r != nil {
			latency = 0
			err = fmt.Errorf("panic in MihomoDelay: %v", r)
		}
	}()

	proxyAdapter, err := GetMihomoAdapter(nodeLink)
	if err != nil {
		return 0, err
	}

	if testUrl == "" {
		testUrl = "http://cp.cloudflare.com/generate_204"
	}

	parsedUrl, err := url.Parse(testUrl)
	if err != nil {
		return 0, fmt.Errorf("parse test url error: %v", err)
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
		return 0, fmt.Errorf("invalid port: %v", err)
	}
	// Validate port range to prevent overflow
	if portInt < 0 || portInt > 65535 {
		return 0, fmt.Errorf("port out of range: %d", portInt)
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
		return 0, fmt.Errorf("dial error: %v", err)
	}
	// Close connection asynchronously to avoid blocking if it hangs
	defer func() {
		go func() {
			_ = conn.Close()
		}()
	}()

	latency = int(time.Since(start).Milliseconds())
	return latency, nil
}

// MihomoSpeedTest performs a true speed test using Mihomo adapter
// Returns speed in MB/s and latency in ms
func MihomoSpeedTest(nodeLink string, testUrl string, timeout time.Duration) (speed float64, latency int, err error) {
	// Recover from any panics and return error with zero values
	defer func() {
		if r := recover(); r != nil {
			speed = 0
			latency = 0
			err = fmt.Errorf("panic in MihomoSpeedTest: %v", r)
		}
	}()

	proxyAdapter, err := GetMihomoAdapter(nodeLink)
	if err != nil {
		return 0, 0, err
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
	// Validate port range to prevent overflow
	if portInt < 0 || portInt > 65535 {
		return 0, 0, fmt.Errorf("port out of range: %d", portInt)
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
		return 0, latency, fmt.Errorf("create request error: %v", err)
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
		return 0, latency, fmt.Errorf("http get error: %v", err)
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
			return 0, latency, fmt.Errorf("read body error: %v", err)
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
		return 0, latency, nil
	}

	// Speed in MB/s
	speed = float64(totalRead) / 1024 / 1024 / duration.Seconds()

	return speed, latency, nil
}
