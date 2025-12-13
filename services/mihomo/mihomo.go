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
	proxyStruct, err := protocol.LinkToProxy(protocol.Urls{Url: nodeLink}, sqlConfig)
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
// Returns latency in ms by measuring actual HTTP request round-trip time
// This is the legacy function that includes handshake time by default
func MihomoDelay(nodeLink string, testUrl string, timeout time.Duration) (latency int, err error) {
	proxyAdapter, err := GetMihomoAdapter(nodeLink)
	if err != nil {
		return 0, err
	}
	return MihomoDelayWithAdapter(proxyAdapter, testUrl, timeout)
}

// MihomoDelayWithAdapter performs latency test with an existing adapter (avoids repeated adapter creation)
// Returns latency in ms by measuring actual HTTP request round-trip time including handshake
func MihomoDelayWithAdapter(proxyAdapter constant.Proxy, testUrl string, timeout time.Duration) (latency int, err error) {
	// Recover from any panics and return error with zero latency
	defer func() {
		if r := recover(); r != nil {
			latency = 0
			err = fmt.Errorf("panic in MihomoDelayWithAdapter: %v", r)
		}
	}()

	if testUrl == "" {
		testUrl = "http://cp.cloudflare.com/generate_204"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create HTTP client with custom transport that uses the proxy adapter
	// DisableKeepAlives: true ensures fresh connection for accurate measurement
	client := createHTTPClientWithAdapter(proxyAdapter, timeout, true)

	// Use HEAD request for minimal data transfer, more accurate latency measurement
	req, err := http.NewRequestWithContext(ctx, "HEAD", testUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("create request error: %v", err)
	}

	// Measure actual HTTP round-trip time
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http request error: %v", err)
	}
	latency = int(time.Since(start).Milliseconds())

	// Clean up response body
	if resp != nil && resp.Body != nil {
		go func() {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}()
	}

	return latency, nil
}

// createHTTPClientWithAdapter creates an HTTP client using the proxy adapter
// disableKeepAlives: true for measuring full connection time, false for reusing connections
func createHTTPClientWithAdapter(proxyAdapter constant.Proxy, timeout time.Duration, disableKeepAlives bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Recover from panics in DialContext
				var dialErr error
				defer func() {
					if r := recover(); r != nil {
						dialErr = fmt.Errorf("panic in DialContext: %v", r)
					}
				}()

				// Parse addr to get host and port for metadata
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
				conn, err := proxyAdapter.DialContext(ctx, md)
				if dialErr != nil {
					return nil, dialErr
				}
				return conn, err
			},
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: disableKeepAlives,
		},
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects for latency test
		},
	}
}

// MihomoDelayWithSamples performs multiple latency tests and returns the average
// It removes the highest outlier to improve accuracy
// samples: number of samples to take (minimum 1, recommended 3)
// includeHandshake: if true, measure full connection time; if false, warmup first then measure pure RTT
func MihomoDelayWithSamples(nodeLink string, testUrl string, timeout time.Duration, samples int, includeHandshake bool) (latency int, err error) {
	// Recover from any panics
	defer func() {
		if r := recover(); r != nil {
			latency = 0
			err = fmt.Errorf("panic in MihomoDelayWithSamples: %v", r)
		}
	}()

	if samples < 1 {
		samples = 1
	}
	if samples > 10 {
		samples = 10 // Cap at 10 to prevent abuse
	}

	if testUrl == "" {
		testUrl = "http://cp.cloudflare.com/generate_204"
	}

	// Create adapter once and reuse for all samples (CPU optimization)
	proxyAdapter, err := GetMihomoAdapter(nodeLink)
	if err != nil {
		return 0, err
	}

	var results []int
	var lastErr error

	if includeHandshake {
		// Mode 1: Include handshake time - each sample uses fresh connection
		for i := 0; i < samples; i++ {
			lat, sampleErr := MihomoDelayWithAdapter(proxyAdapter, testUrl, timeout)
			if sampleErr != nil {
				lastErr = sampleErr
				continue
			}
			results = append(results, lat)
		}
	} else {
		// Mode 2: Exclude handshake - warmup first, then measure pure RTT
		// Create client with keep-alive enabled for connection reuse
		client := createHTTPClientWithAdapter(proxyAdapter, timeout, false)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Warmup request - establish connection (don't measure this)
		warmupReq, err := http.NewRequestWithContext(ctx, "HEAD", testUrl, nil)
		if err != nil {
			return 0, fmt.Errorf("create warmup request error: %v", err)
		}
		warmupResp, warmupErr := client.Do(warmupReq)
		if warmupErr != nil {
			// Warmup failed - node is unreachable, return immediately
			return 0, fmt.Errorf("warmup connection failed: %v", warmupErr)
		}
		if warmupResp != nil && warmupResp.Body != nil {
			_, _ = io.Copy(io.Discard, warmupResp.Body)
			_ = warmupResp.Body.Close()
		}

		// Now measure pure RTT using the warmed-up connection
		for i := 0; i < samples; i++ {
			sampleCtx, sampleCancel := context.WithTimeout(context.Background(), timeout)
			req, reqErr := http.NewRequestWithContext(sampleCtx, "HEAD", testUrl, nil)
			if reqErr != nil {
				sampleCancel()
				lastErr = reqErr
				continue
			}

			start := time.Now()
			resp, doErr := client.Do(req)
			lat := int(time.Since(start).Milliseconds())
			sampleCancel()

			if doErr != nil {
				lastErr = doErr
				continue
			}
			if resp != nil && resp.Body != nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}
			results = append(results, lat)
		}
	}

	if len(results) == 0 {
		if lastErr != nil {
			return 0, lastErr
		}
		return 0, fmt.Errorf("all samples failed")
	}

	// If only one result, return it directly
	if len(results) == 1 {
		return results[0], nil
	}

	// Remove highest outlier if we have more than 2 samples
	if len(results) > 2 {
		maxIdx := 0
		for i, v := range results {
			if v > results[maxIdx] {
				maxIdx = i
			}
		}
		results = append(results[:maxIdx], results[maxIdx+1:]...)
	}

	// Calculate average
	sum := 0
	for _, v := range results {
		sum += v
	}
	return sum / len(results), nil
}

// MihomoSpeedTest performs a true speed test using Mihomo adapter
// Returns speed in MB/s, latency in ms, and bytes downloaded
func MihomoSpeedTest(nodeLink string, testUrl string, timeout time.Duration) (speed float64, latency int, bytesDownloaded int64, err error) {
	// Recover from any panics and return error with zero values
	defer func() {
		if r := recover(); r != nil {
			speed = 0
			latency = 0
			bytesDownloaded = 0
			err = fmt.Errorf("panic in MihomoSpeedTest: %v", r)
		}
	}()

	proxyAdapter, err := GetMihomoAdapter(nodeLink)
	if err != nil {
		return 0, 0, 0, err
	}

	// 4. Perform Speed Test
	// We will try to download from testUrl
	if testUrl == "" {
		testUrl = "https://speed.cloudflare.com/__down?bytes=10000000" // Default 10MB
	}

	parsedUrl, err := url.Parse(testUrl)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse test url error: %v", err)
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
		return 0, 0, 0, fmt.Errorf("invalid port: %v", err)
	}
	// Validate port range to prevent overflow
	if portInt < 0 || portInt > 65535 {
		return 0, 0, 0, fmt.Errorf("port out of range: %d", portInt)
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
		return 0, 0, 0, fmt.Errorf("dial error: %v", err)
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
		return 0, latency, 0, fmt.Errorf("create request error: %v", err)
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
		return 0, latency, 0, fmt.Errorf("http get error: %v", err)
	}
	defer resp.Body.Close()

	// Read body to measure speed
	// We can read up to N bytes or until EOF
	buf := make([]byte, 32*1024)
	var totalRead int64 // Changed to int64 to avoid overflow for large downloads
	readStart := time.Now()

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
			return 0, latency, totalRead, fmt.Errorf("read body error: %v", err)
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
		return 0, latency, totalRead, nil
	}

	// Speed in MB/s
	speed = float64(totalRead) / 1024 / 1024 / duration.Seconds()

	return speed, latency, totalRead, nil
}

// FetchLandingIP fetches the landing IP address through the proxy by accessing https://api.ip.sb/ip
// Returns the IP address as a string
func FetchLandingIP(nodeLink string, timeout time.Duration) (string, error) {
	// Recover from any panics and return error
	defer func() {
		if r := recover(); r != nil {
			// Don't do anything, just return from the function
		}
	}()

	proxyAdapter, err := GetMihomoAdapter(nodeLink)
	if err != nil {
		return "", err
	}

	testUrl := "https://api.ip.sb/ip"
	parsedUrl, err := url.Parse(testUrl)
	if err != nil {
		return "", fmt.Errorf("parse test url error: %v", err)
	}

	port := uint16(443)

	metadata := &constant.Metadata{
		Host:    parsedUrl.Hostname(),
		DstPort: port,
		Type:    constant.HTTP,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := proxyAdapter.DialContext(ctx, metadata)
	if err != nil {
		return "", fmt.Errorf("dial error: %v", err)
	}
	defer func() {
		go func() {
			_ = conn.Close()
		}()
	}()

	// Create HTTP client with the proxy transport
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic in DialContext: %v", r)
					}
				}()

				h, pStr, splitErr := net.SplitHostPort(addr)
				if splitErr != nil {
					return nil, fmt.Errorf("split host port error: %v", splitErr)
				}

				pInt, atoiErr := strconv.Atoi(pStr)
				if atoiErr != nil {
					return nil, fmt.Errorf("invalid port string: %v", atoiErr)
				}

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
		return "", fmt.Errorf("http get error: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body (should be just the IP address)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body error: %v", err)
	}

	// Trim whitespace and newlines
	ip := strings.TrimSpace(string(body))

	return ip, nil
}
