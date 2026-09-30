package backup

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	maxWebDAVResponseBytes = int64(2 << 20)
	MaxBackupArchiveBytes  = int64(1 << 30)
)

var blockedWebDAVPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("3fff::/20"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

// RemoteFile describes one backup ZIP stored in WebDAV.
type RemoteFile struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// Client is a minimal WebDAV client for SublinkPro backup operations.
type Client struct {
	config  Config
	baseURL *url.URL
	http    *http.Client
}

func NewClient(cfg Config) (*Client, error) {
	cfg, err := normalizeAndValidateConfig(cfg, true)
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 WebDAV 地址失败: %w", err)
	}
	dialContext := safeDialContext(cfg.AllowPrivateNetwork)
	transport := &http.Transport{
		DialContext:           dialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: cfg.Timeout(),
	}
	return &Client{
		config:  cfg,
		baseURL: baseURL,
		http: &http.Client{
			Timeout:   cfg.Timeout(),
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 0 && strings.EqualFold(via[0].Method, "MKCOL") {
					// Many WebDAV servers, including TeraCLOUD, canonicalize collections
					// with a trailing slash. Do not convert MKCOL into GET.
					return http.ErrUseLastResponse
				}
				return errors.New("WebDAV 重定向已被拒绝")
			},
		},
	}, nil
}

func safeDialContext(allowPrivate bool) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if allowPrivate {
			return dialer.DialContext(ctx, network, address)
		}
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("解析 WebDAV 目标地址失败: %w", err)
		}
		addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("解析 WebDAV 主机失败: %w", err)
		}
		if len(addresses) == 0 {
			return nil, errors.New("WebDAV 主机没有可用地址")
		}
		for _, resolved := range addresses {
			if !isPublicWebDAVIP(resolved.IP) {
				return nil, fmt.Errorf("WebDAV 主机解析到私有或保留地址 %s；如需连接内网 NAS，请显式允许私有网络", resolved.IP)
			}
		}
		var lastErr error
		for _, resolved := range addresses {
			conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(resolved.IP.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}
		return nil, lastErr
	}
}

func isPublicWebDAVIP(ip net.IP) bool {
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	if !address.IsGlobalUnicast() {
		return false
	}
	for _, prefix := range blockedWebDAVPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func (c *Client) Test(ctx context.Context) error {
	resp, err := c.request(ctx, "PROPFIND", c.directoryURL(), strings.NewReader(propfindBody), map[string]string{"Depth": "0", "Content-Type": "application/xml"}, -1)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMultiStatus {
		return c.statusError(resp, "测试 WebDAV 连接")
	}
	body, err := readLimited(resp.Body, maxWebDAVResponseBytes)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	var result davMultiStatus
	if err := xml.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析 WebDAV 连接测试响应失败: %w", err)
	}
	for _, response := range result.Responses {
		if response.successful() {
			return nil
		}
	}
	return errors.New("WebDAV 连接测试未返回可访问的目录")
}

func (c *Client) EnsureDirectory(ctx context.Context) error {
	segments := strings.Split(c.config.RemotePath, "/")
	for index := range segments {
		target := strings.TrimRight(c.urlForSegments(segments[:index+1]...), "/") + "/"
		resp, err := c.request(ctx, "MKCOL", target, nil, nil, 0)
		if err != nil {
			return err
		}
		status := resp.StatusCode
		accepted := isSuccessfulMKCOL(resp, target)
		_ = resp.Body.Close()
		if accepted {
			continue
		}
		if status == http.StatusConflict {
			return fmt.Errorf("创建 WebDAV 目录失败，父目录不存在或无权限 (HTTP %d)", status)
		}
		return fmt.Errorf("创建 WebDAV 目录失败 (HTTP %d)", status)
	}
	return nil
}

func (c *Client) Upload(ctx context.Context, archive Archive) (RemoteFile, error) {
	if archive.Size <= 0 || archive.Size > MaxBackupArchiveBytes {
		return RemoteFile{}, fmt.Errorf("备份大小必须在 1 字节到 %d 字节之间", MaxBackupArchiveBytes)
	}
	if err := validateBackupFilename(archive.Name); err != nil {
		return RemoteFile{}, err
	}
	if err := c.EnsureDirectory(ctx); err != nil {
		return RemoteFile{}, err
	}
	file, err := openArchive(archive.Path)
	if err != nil {
		return RemoteFile{}, err
	}
	defer func() { _ = file.Close() }()

	resp, err := c.request(ctx, http.MethodPut, c.fileURL(archive.Name), io.LimitReader(file, MaxBackupArchiveBytes+1), map[string]string{"Content-Type": "application/zip"}, archive.Size)
	if err != nil {
		return RemoteFile{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return RemoteFile{}, c.statusError(resp, "上传 WebDAV 备份")
	}
	return RemoteFile{Name: archive.Name, Size: archive.Size, ModifiedAt: archive.Modified}, nil
}

func (c *Client) List(ctx context.Context) ([]RemoteFile, error) {
	resp, err := c.request(ctx, "PROPFIND", c.directoryURL(), strings.NewReader(propfindBody), map[string]string{"Depth": "1", "Content-Type": "application/xml"}, -1)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMultiStatus {
		return nil, c.statusError(resp, "读取 WebDAV 备份列表")
	}
	body, err := readLimited(resp.Body, maxWebDAVResponseBytes)
	if err != nil {
		return nil, err
	}
	var result davMultiStatus
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 WebDAV 文件列表失败: %w", err)
	}
	files := make([]RemoteFile, 0, len(result.Responses))
	for _, response := range result.Responses {
		item, ok := response.remoteFile()
		if ok {
			files = append(files, item)
		}
	}
	return files, nil
}

func (c *Client) Download(ctx context.Context, filename string, dst io.Writer) (int64, error) {
	if err := validateBackupFilename(filename); err != nil {
		return 0, err
	}
	resp, err := c.request(ctx, http.MethodGet, c.fileURL(filename), nil, nil, -1)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, c.statusError(resp, "下载 WebDAV 备份")
	}
	if resp.ContentLength > MaxBackupArchiveBytes {
		return 0, fmt.Errorf("远程备份超过 %d 字节限制", MaxBackupArchiveBytes)
	}
	written, err := io.Copy(dst, io.LimitReader(resp.Body, MaxBackupArchiveBytes+1))
	if err != nil {
		return written, fmt.Errorf("写入 WebDAV 备份失败: %w", err)
	}
	if written > MaxBackupArchiveBytes {
		return written, fmt.Errorf("远程备份超过 %d 字节限制", MaxBackupArchiveBytes)
	}
	return written, nil
}

func (c *Client) request(ctx context.Context, method, target string, body io.Reader, headers map[string]string, contentLength int64) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, fmt.Errorf("创建 WebDAV 请求失败: %w", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if contentLength >= 0 {
		req.ContentLength = contentLength
	}
	if c.config.Username != "" || c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WebDAV 请求失败: %w", err)
	}
	return resp, nil
}

func (c *Client) statusError(resp *http.Response, action string) error {
	_, _ = readLimited(resp.Body, 4096)
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("%s失败：用户名或密码错误", action)
	case http.StatusForbidden:
		return fmt.Errorf("%s失败：没有远程目录权限", action)
	case http.StatusNotFound:
		return fmt.Errorf("%s失败：远程路径不存在", action)
	default:
		return fmt.Errorf("%s失败 (HTTP %d)", action, resp.StatusCode)
	}
}

func isSuccessfulMKCOL(resp *http.Response, target string) bool {
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK, http.StatusNoContent, http.StatusMethodNotAllowed:
		return true
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		// Some servers canonicalize collections with a redirect. Never follow it or
		// treat a redirect to another origin as a successful directory creation.
		location := resp.Header.Get("Location")
		if location == "" {
			return false
		}
		original, err := url.Parse(target)
		if err != nil {
			return false
		}
		redirected, err := original.Parse(location)
		return err == nil && strings.EqualFold(original.Scheme, redirected.Scheme) && strings.EqualFold(original.Host, redirected.Host)
	default:
		return false
	}
}

func (c *Client) directoryURL() string {
	return strings.TrimRight(c.urlForSegments(strings.Split(c.config.RemotePath, "/")...), "/") + "/"
}

func (c *Client) fileURL(filename string) string {
	segments := append(strings.Split(c.config.RemotePath, "/"), filename)
	return c.urlForSegments(segments...)
}

func (c *Client) urlForSegments(segments ...string) string {
	clone := *c.baseURL
	joined := clone.Path
	for _, segment := range segments {
		joined = path.Join(joined, segment)
	}
	clone.Path = joined
	clone.RawPath = ""
	return clone.String()
}

func validateBackupFilename(filename string) error {
	if filename == "" || filename != path.Base(filename) || strings.Contains(filename, "\\") || strings.ContainsRune(filename, '\x00') {
		return errors.New("WebDAV 备份文件名无效")
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".zip") {
		return errors.New("仅支持 ZIP 备份文件")
	}
	return nil
}

func readLimited(reader io.Reader, limit int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("WebDAV 响应超过 %d 字节限制", limit)
	}
	return body, nil
}

var openArchive = func(filePath string) (io.ReadCloser, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开备份文件失败: %w", err)
	}
	return file, nil
}

const propfindBody = `<?xml version="1.0" encoding="utf-8" ?>
<d:propfind xmlns:d="DAV:">
  <d:prop><d:resourcetype/><d:getcontentlength/><d:getlastmodified/></d:prop>
</d:propfind>`

type davMultiStatus struct {
	Responses []davResponse `xml:"response"`
}

type davResponse struct {
	Href      string        `xml:"href"`
	Status    string        `xml:"status"`
	PropStats []davPropStat `xml:"propstat"`
}

type davPropStat struct {
	Status string  `xml:"status"`
	Prop   davProp `xml:"prop"`
}

type davProp struct {
	ContentLength string          `xml:"getcontentlength"`
	LastModified  string          `xml:"getlastmodified"`
	ResourceType  davResourceType `xml:"resourcetype"`
}

type davResourceType struct {
	Collection *struct{} `xml:"collection"`
}

func (response davResponse) successful() bool {
	if strings.Contains(response.Status, " 200 ") {
		return true
	}
	for _, propStat := range response.PropStats {
		if strings.Contains(propStat.Status, " 200 ") || propStat.Status == "" {
			return true
		}
	}
	return false
}

func (response davResponse) remoteFile() (RemoteFile, bool) {
	parsed, err := url.Parse(strings.TrimSpace(response.Href))
	if err != nil {
		return RemoteFile{}, false
	}
	name := path.Base(strings.TrimSuffix(parsed.Path, "/"))
	if err := validateBackupFilename(name); err != nil {
		return RemoteFile{}, false
	}
	for _, propStat := range response.PropStats {
		if propStat.Prop.ResourceType.Collection != nil || (!strings.Contains(propStat.Status, " 200 ") && propStat.Status != "") {
			continue
		}
		size, _ := strconv.ParseInt(strings.TrimSpace(propStat.Prop.ContentLength), 10, 64)
		modified, _ := http.ParseTime(strings.TrimSpace(propStat.Prop.LastModified))
		return RemoteFile{Name: name, Size: size, ModifiedAt: modified}, true
	}
	return RemoteFile{}, false
}
