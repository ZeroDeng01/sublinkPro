package protocol

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"sublink/internal/yamlutil"
	"unicode"

	"github.com/metacubex/mihomo/transport/sudoku"
	"gopkg.in/yaml.v3"
)

const sudokuURLPrefix = "sudoku://"

var sudokuYAMLFields = yamlFieldNames(reflect.TypeFor[SudokuProxyOptions]())

type sudokuProtocol struct{ *ProxyProtocolSpec }

func init() {
	base := newProtocolSpec("sudoku", []string{sudokuURLPrefix}, "Sudoku", "#455a64", "S", Sudoku{}, "Name", DecodeSudokuURL, EncodeSudokuURL, func(s Sudoku) LinkIdentity {
		return buildIdentity("sudoku", s.Name, s.Server, strconv.Itoa(s.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "Key", Label: "客户端密钥", Type: "string", Group: "auth", Secret: true},
		FieldMeta{Name: "AEADMethod", Label: "加密方式", Type: "string", Group: "transport", Options: []string{"chacha20-poly1305", "aes-128-gcm", "none"}},
		FieldMeta{Name: "TableType", Label: "编码表类型", Type: "string", Group: "transport", Options: []string{"prefer_ascii", "prefer_entropy", "up_ascii_down_entropy", "up_entropy_down_ascii"}},
		FieldMeta{Name: "PackedDownlink", Label: "带宽优化下行", Type: "bool", Group: "transport"},
		FieldMeta{Name: "PaddingMin", Label: "最小填充率", Type: "int", Group: "advanced", Advanced: true},
		FieldMeta{Name: "PaddingMax", Label: "最大填充率", Type: "int", Group: "advanced", Advanced: true},
		FieldMeta{Name: "CustomTable", Label: "自定义编码表", Type: "string", Group: "advanced", Advanced: true},
		FieldMeta{Name: "CustomTables", Label: "自定义编码表列表", Type: "string", Group: "advanced", Advanced: true, Placeholder: "xpxvvpvv, vxpvxvvp"},
		FieldMeta{Name: "Multiplex", Label: "多路复用", Type: "string", Group: "transport", Options: []string{"off", "auto", "on"}},
		FieldMeta{Name: "DisableHTTPMask", Label: "禁用 HTTPMask", Type: "bool", Group: "transport"},
		FieldMeta{Name: "HTTPMaskMode", Label: "HTTPMask 模式", Type: "string", Group: "transport", Options: []string{"legacy", "stream", "poll", "auto", "ws"}},
		FieldMeta{Name: "HTTPMaskTLS", Label: "HTTPMask TLS", Type: "bool", Group: "tls"},
		FieldMeta{Name: "HTTPMaskHost", Label: "HTTPMask Host/SNI", Type: "string", Group: "transport"},
		FieldMeta{Name: "HTTPMaskPathRoot", Label: "HTTPMask 路径前缀", Type: "string", Group: "transport"},
	).WithClientSupport(ClientClash, ClientMihomo)

	registered := newProxyProtocolSpec(base, buildSudokuProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "sudoku")
	}, convertProxyToSudoku, EncodeSudokuURL)
	registered.fromProxy = func(proxy Proxy) (string, error) {
		decoded := convertProxyToSudoku(proxy)
		if err := decoded.normalize(); err != nil {
			return "", err
		}
		return EncodeSudokuURL(decoded), nil
	}
	MustRegisterProtocol(sudokuProtocol{registered})
}

// Sudoku represents the official short link. Descriptive JSON keys retain
// Mihomo-only settings when SublinkPro stores a node imported from YAML.
type Sudoku struct {
	Name             string `json:"n,omitempty"`
	Server           string `json:"h"`
	Port             int    `json:"p"`
	Key              string `json:"k"`
	TableType        string `json:"a,omitempty"`
	AEADMethod       string `json:"e,omitempty"`
	LocalPort        int    `json:"m,omitempty"`
	PackedDownlink   bool   `json:"x,omitempty"`
	CustomTable      string `json:"t,omitempty"`
	CustomTables     string `json:"-"` // Comma-separated for the raw node editor.
	DisableHTTPMask  bool   `json:"hd,omitempty"`
	HTTPMaskMode     string `json:"hm,omitempty"`
	HTTPMaskTLS      bool   `json:"ht,omitempty"`
	HTTPMaskHost     string `json:"hh,omitempty"`
	Multiplex        string `json:"hx,omitempty"`
	HTTPMaskPathRoot string `json:"hy,omitempty"`
	PaddingMin       int    `json:"padding-min"`
	PaddingMax       int    `json:"padding-max"`
	UDP              bool   `json:"udp,omitempty"`
	TFO              bool   `json:"tfo,omitempty"`
	MPTCP            bool   `json:"mptcp,omitempty"`
	InterfaceName    string `json:"interface-name,omitempty"`
	RoutingMark      int    `json:"routing-mark,omitempty"`
	IPVersion        string `json:"ip-version,omitempty"`
	DialerProxy      string `json:"dialer-proxy,omitempty"`
}

type sudokuLinkFields Sudoku

type sudokuLinkPayload struct {
	sudokuLinkFields
	CustomTables []string `json:"ts,omitempty"`
}

type SudokuHTTPMask struct {
	Disable   bool   `yaml:"disable,omitempty"`
	Mode      string `yaml:"mode,omitempty"`
	TLS       bool   `yaml:"tls,omitempty"`
	Host      string `yaml:"host,omitempty"`
	PathRoot  string `yaml:"path-root,omitempty"`
	Multiplex string `yaml:"multiplex,omitempty"`
}

// SudokuProxyOptions contains Mihomo's Sudoku fields and its legacy flat aliases.
type SudokuProxyOptions struct {
	AEADMethod         string          `yaml:"aead-method,omitempty"`
	PaddingMin         *int            `yaml:"padding-min,omitempty"`
	PaddingMax         *int            `yaml:"padding-max,omitempty"`
	TableType          string          `yaml:"table-type,omitempty"`
	CustomTable        string          `yaml:"custom-table,omitempty"`
	CustomTables       []string        `yaml:"custom-tables,omitempty"`
	EnablePureDownlink *bool           `yaml:"enable-pure-downlink,omitempty"`
	Multiplex          string          `yaml:"multiplex,omitempty"`
	HTTPMask           *SudokuHTTPMask `yaml:"httpmask,omitempty"`
	HTTPMaskEnabled    *bool           `yaml:"http-mask,omitempty"`
	HTTPMaskMode       string          `yaml:"http-mask-mode,omitempty"`
	HTTPMaskTLS        bool            `yaml:"http-mask-tls,omitempty"`
	HTTPMaskHost       string          `yaml:"http-mask-host,omitempty"`
	HTTPMaskMultiplex  string          `yaml:"http-mask-multiplex,omitempty"`
	HTTPMaskPathRoot   string          `yaml:"path-root,omitempty"`
}

func (sudokuProtocol) decodeProxyYAML(value *yaml.Node, proxy *Proxy) error {
	options := &SudokuProxyOptions{}
	if err := yamlutil.DecodeWithStringBools(value, options); err != nil {
		return err
	}
	proxy.SudokuOptions = options
	return nil
}

func (sudokuProtocol) encodeProxyYAML(proxy Proxy) any {
	return struct {
		proxyYAML           `yaml:",inline"`
		*SudokuProxyOptions `yaml:",inline"`
	}{proxyYAML(proxy), proxy.SudokuOptions}
}

func (sudokuProtocol) proxyYAMLFieldNames() map[string]struct{} {
	return sudokuYAMLFields
}

func (sudokuProtocol) normalizeProxyForHash(proxy Proxy) Proxy {
	// Compare effective settings when YAML and short links use different aliases or defaults.
	node := convertProxyToSudoku(proxy)
	if node.normalize() != nil {
		return proxy
	}
	normalized := node.toProxy()
	normalized.ClashExtra = proxy.ClashExtra
	return normalized
}

func DecodeSudokuURL(link string) (Sudoku, error) {
	scheme, payload, ok := strings.Cut(strings.TrimSpace(link), "://")
	if !ok || !strings.EqualFold(scheme, "sudoku") {
		return Sudoku{}, errors.New("invalid Sudoku link scheme")
	}
	encoding := base64.RawURLEncoding
	if strings.HasSuffix(payload, "=") {
		encoding = base64.URLEncoding
	}
	data, err := encoding.DecodeString(payload)
	if err != nil {
		return Sudoku{}, errors.New("invalid Sudoku short link encoding")
	}

	// Match the standalone short-link reader's defaults.
	decoded := sudokuLinkPayload{sudokuLinkFields: sudokuLinkFields{PaddingMin: 5, PaddingMax: 15, LocalPort: 1080}}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return Sudoku{}, errors.New("invalid Sudoku short link payload")
	}
	node := Sudoku(decoded.sudokuLinkFields)
	node.CustomTables = strings.Join(decoded.CustomTables, ", ")
	if err := node.normalize(); err != nil {
		return Sudoku{}, err
	}
	return node, nil
}

func EncodeSudokuURL(node Sudoku) string {
	switch node.TableType {
	case "prefer_ascii":
		node.TableType = "ascii"
	case "prefer_entropy":
		node.TableType = "entropy"
	}
	// Sudoku contains only JSON-compatible scalar values and strings.
	data, _ := json.Marshal(sudokuLinkPayload{
		sudokuLinkFields: sudokuLinkFields(node),
		CustomTables:     splitSudokuTables(node.CustomTables),
	})
	return sudokuURLPrefix + base64.RawURLEncoding.EncodeToString(data)
}

func splitSudokuTables(tables string) []string {
	return strings.FieldsFunc(tables, func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
}

func (s *Sudoku) normalize() error {
	s.Server = strings.TrimSpace(s.Server)
	if strings.HasPrefix(s.Server, "[") && strings.HasSuffix(s.Server, "]") {
		s.Server = s.Server[1 : len(s.Server)-1]
	}
	if s.Server == "" || strings.ContainsAny(s.Server, "/?#@[],") || strings.ContainsFunc(s.Server, unicode.IsSpace) {
		return errors.New("invalid Sudoku server host")
	}
	if strings.Contains(s.Server, ":") {
		if address, err := netip.ParseAddr(s.Server); err != nil || !address.Is6() {
			return errors.New("invalid Sudoku IPv6 address")
		}
	}
	if s.Port < 1 || s.Port > 65535 || strings.TrimSpace(s.Key) == "" {
		return errors.New("Sudoku requires a valid server port and a client key")
	}
	if s.PaddingMin < 0 || s.PaddingMax > 100 || s.PaddingMax < s.PaddingMin {
		return errors.New("Sudoku padding must satisfy 0 <= padding-min <= padding-max <= 100")
	}
	var err error
	if s.TableType, err = sudoku.NormalizeTableType(s.TableType); err != nil {
		return err
	}
	switch s.AEADMethod {
	case "":
		s.AEADMethod = "none"
	case "chacha20-poly1305", "aes-128-gcm", "none":
	default:
		return errors.New("invalid Sudoku AEAD method")
	}
	s.HTTPMaskMode = strings.ToLower(strings.TrimSpace(s.HTTPMaskMode))
	switch s.HTTPMaskMode {
	case "":
		s.HTTPMaskMode = "legacy"
	case "legacy", "stream", "poll", "auto", "ws":
	default:
		return errors.New("invalid Sudoku HTTPMask mode")
	}
	if s.Multiplex, err = sudoku.NormalizeMultiplexMode(s.Multiplex); err != nil {
		return errors.New("invalid Sudoku multiplex mode")
	}
	if s.LocalPort == 0 {
		s.LocalPort = 1080
	}
	if strings.TrimSpace(s.Name) == "" {
		s.Name = formatURLHostPort(s.Server, strconv.Itoa(s.Port))
	}
	return nil
}

func buildSudokuProxy(link Urls, config OutputConfig) (Proxy, error) {
	node, err := DecodeSudokuURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	proxy := node.toProxy()
	proxy.Udp = proxy.Udp || config.Udp
	if link.DialerProxyName != "" {
		proxy.Dialer_proxy = link.DialerProxyName
	}
	return proxy, nil
}

func (s Sudoku) toProxy() Proxy {
	pureDownlink := !s.PackedDownlink
	return Proxy{
		Name: s.Name, Type: "sudoku", Server: s.Server, Port: FlexPort(s.Port), Key: s.Key,
		SudokuOptions: &SudokuProxyOptions{
			AEADMethod: s.AEADMethod, TableType: s.TableType, PaddingMin: &s.PaddingMin, PaddingMax: &s.PaddingMax,
			EnablePureDownlink: &pureDownlink, CustomTable: s.CustomTable, CustomTables: splitSudokuTables(s.CustomTables),
			Multiplex: s.Multiplex,
			HTTPMask: &SudokuHTTPMask{
				Disable: s.DisableHTTPMask, Mode: s.HTTPMaskMode, TLS: s.HTTPMaskTLS, Host: s.HTTPMaskHost, PathRoot: s.HTTPMaskPathRoot,
			},
		},
		Udp: s.UDP, Tfo: s.TFO, Mptcp: s.MPTCP, Interface_name: s.InterfaceName,
		Routing_mark: s.RoutingMark, Ip_version: s.IPVersion, Dialer_proxy: s.DialerProxy,
	}
}

func convertProxyToSudoku(proxy Proxy) Sudoku {
	options := proxy.SudokuOptions
	if options == nil {
		options = &SudokuProxyOptions{}
	}
	defaults := sudoku.DefaultConfig()
	paddingMin, paddingMax := sudoku.ResolvePadding(options.PaddingMin, options.PaddingMax, defaults.PaddingMin, defaults.PaddingMax)
	node := Sudoku{
		Name: proxy.Name, Server: proxy.Server, Port: proxy.Port.Int(), Key: proxy.Key,
		AEADMethod: options.AEADMethod, TableType: options.TableType, PaddingMin: paddingMin, PaddingMax: paddingMax,
		PackedDownlink: !sudoku.DerefBool(options.EnablePureDownlink, defaults.EnablePureDownlink),
		CustomTable:    options.CustomTable, CustomTables: strings.Join(options.CustomTables, ", "),
		DisableHTTPMask: !sudoku.DerefBool(options.HTTPMaskEnabled, !defaults.DisableHTTPMask),
		HTTPMaskMode:    options.HTTPMaskMode, HTTPMaskTLS: options.HTTPMaskTLS, HTTPMaskHost: options.HTTPMaskHost,
		HTTPMaskPathRoot: options.HTTPMaskPathRoot, Multiplex: options.Multiplex,
		UDP: proxy.Udp, TFO: proxy.Tfo, MPTCP: proxy.Mptcp, InterfaceName: proxy.Interface_name,
		RoutingMark: proxy.Routing_mark, IPVersion: proxy.Ip_version, DialerProxy: proxy.Dialer_proxy,
	}
	if node.AEADMethod == "" {
		node.AEADMethod = defaults.AEADMethod
	}
	if mode := strings.TrimSpace(options.HTTPMaskMultiplex); mode != "" {
		node.Multiplex = mode
	}
	// Apply nested HTTPMask options in the same order as the Mihomo adapter.
	if mask := options.HTTPMask; mask != nil {
		node.DisableHTTPMask, node.HTTPMaskTLS, node.HTTPMaskHost = mask.Disable, mask.TLS, mask.Host
		if mask.Mode != "" {
			node.HTTPMaskMode = mask.Mode
		}
		if path := strings.TrimSpace(mask.PathRoot); path != "" {
			node.HTTPMaskPathRoot = path
		}
		if mode := strings.TrimSpace(mask.Multiplex); mode != "" {
			node.Multiplex = mode
		}
	}
	return node
}
