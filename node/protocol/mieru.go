package protocol

import (
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sublink/utils"

	"github.com/enfein/mieru/v3/apis/trafficpattern"
	pb "github.com/enfein/mieru/v3/pkg/appctl/appctlpb"
	"google.golang.org/protobuf/proto"
)

func init() {
	base := newProtocolSpec("mieru", []string{"mieru://", "mierus://"}, "Mieru", "#5f6caf", "M", Mieru{}, "Name", DecodeMieruURL, EncodeMieruURL, func(m Mieru) LinkIdentity {
		return buildIdentity("mieru", m.Name, m.Server, m.portValue())
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "PortRange", Label: "端口范围", Type: "string", Group: "basic", Placeholder: "2090-2099", Advanced: true},
		FieldMeta{Name: "Transport", Label: "传输协议", Type: "string", Group: "transport", Options: []string{"TCP", "UDP"}},
		FieldMeta{Name: "Username", Label: "用户名", Type: "string", Group: "auth"},
		FieldMeta{Name: "Password", Label: "密码", Type: "string", Group: "auth", Secret: true},
		FieldMeta{Name: "Multiplexing", Label: "多路复用", Type: "string", Group: "transport", Options: []string{"MULTIPLEXING_DEFAULT", "MULTIPLEXING_OFF", "MULTIPLEXING_LOW", "MULTIPLEXING_MIDDLE", "MULTIPLEXING_HIGH"}, Advanced: true},
		FieldMeta{Name: "MTU", Label: "MTU", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "HandshakeMode", Label: "握手模式", Type: "string", Group: "transport", Options: []string{"HANDSHAKE_DEFAULT", "HANDSHAKE_STANDARD", "HANDSHAKE_NO_WAIT"}, Advanced: true},
		FieldMeta{Name: "TrafficPattern", Label: "流量模式", Type: "string", Group: "transport", Advanced: true},
	).WithClientSupport(ClientClash, ClientMihomo)
	MustRegisterProtocol(newProxyProtocolSpec(base, buildMieruProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "mieru")
	}, ConvertProxyToMieru, EncodeMieruURL))
}

// Mieru describes the endpoint represented by a SublinkPro node.
type Mieru struct {
	Name           string
	Server         string
	Port           any
	PortRange      string
	Transport      string
	Username       string
	Password       string
	Multiplexing   string
	HandshakeMode  string
	MTU            int
	TrafficPattern string

	otherBindings   []MieruBinding
	simpleQuery     url.Values
	standardConfig  *pb.ClientConfig
	standardProfile int
}

// MieruBinding pairs a port or range with its transport on the same server.
type MieruBinding struct {
	Port      string
	Transport string
}

func (m Mieru) portValue() string {
	if m.PortRange != "" {
		return m.PortRange
	}
	return utils.GetPortString(m.Port)
}

func DecodeMieruURL(link string) (Mieru, error) {
	scheme, payload, _ := strings.Cut(strings.TrimSpace(link), "://")
	switch strings.ToLower(scheme) {
	case "mieru":
		return decodeMieruStandard(payload)
	case "mierus":
		return decodeMieruSimple("mierus://" + payload)
	default:
		return Mieru{}, errors.New("不支持的 Mieru 链接格式")
	}
}

func decodeMieruStandard(payload string) (Mieru, error) {
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(payload)
	}
	config := &pb.ClientConfig{}
	if err != nil || proto.Unmarshal(data, config) != nil {
		return Mieru{}, errors.New("无效的 Mieru 标准分享链接")
	}

	for i, profile := range config.GetProfiles() {
		if active := config.GetActiveProfile(); active != "" && active != profile.GetProfileName() {
			continue
		}
		servers := profile.GetServers()
		if len(servers) == 0 || len(servers[0].GetPortBindings()) == 0 {
			break
		}
		server := servers[0]
		binding := server.GetPortBindings()[0]
		m := Mieru{
			Name: profile.GetProfileName(), Server: server.GetDomainName(),
			Port: int(binding.GetPort()), PortRange: binding.GetPortRange(), Transport: binding.GetProtocol().String(),
			Username: profile.GetUser().GetName(), Password: profile.GetUser().GetPassword(),
			MTU: int(profile.GetMtu()), TrafficPattern: trafficpattern.Encode(profile.GetTrafficPattern()),
			standardConfig: config, standardProfile: i,
		}
		if m.Server == "" {
			m.Server = server.GetIpAddress()
		}
		if mux := profile.GetMultiplexing(); mux != nil && mux.Level != nil {
			m.Multiplexing = mux.GetLevel().String()
		}
		if profile.HandshakeMode != nil {
			m.HandshakeMode = profile.GetHandshakeMode().String()
		}
		if err := m.validate(); err != nil {
			return Mieru{}, err
		}
		for _, binding := range server.GetPortBindings()[1:] {
			port := binding.GetPortRange()
			if port == "" {
				port = strconv.Itoa(int(binding.GetPort()))
			}
			if err := m.addBinding(port, binding.GetProtocol().String()); err != nil {
				return Mieru{}, err
			}
		}
		return m, nil
	}
	return Mieru{}, errors.New("Mieru 标准分享链接缺少可用的配置或端口")
}

func decodeMieruSimple(link string) (Mieru, error) {
	u, err := url.Parse(link)
	if err != nil || u.Hostname() == "" || u.Port() != "" || u.Path != "" {
		return Mieru{}, errors.New("无效的 Mieru 简单分享链接")
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return Mieru{}, errors.New("无效的 Mieru 简单分享链接参数")
	}
	ports, transports := q["port"], q["protocol"]
	if len(ports) == 0 || len(ports) != len(transports) {
		return Mieru{}, errors.New("Mieru 的 port 和 protocol 参数必须成对出现")
	}
	for _, key := range []string{"profile", "mtu", "multiplexing", "handshake-mode", "traffic-pattern"} {
		if len(q[key]) > 1 {
			return Mieru{}, fmt.Errorf("Mieru 参数 %s 只能出现一次", key)
		}
	}
	mtu := 0
	if raw := q.Get("mtu"); raw != "" {
		mtu, err = strconv.Atoi(raw)
		if err != nil {
			return Mieru{}, errors.New("无效的 Mieru MTU")
		}
	}
	password, _ := u.User.Password()
	m := Mieru{
		Name: q.Get("profile"), Server: u.Hostname(), Username: u.User.Username(), Password: password,
		Transport: transports[0], MTU: mtu, Multiplexing: q.Get("multiplexing"),
		HandshakeMode: q.Get("handshake-mode"), TrafficPattern: q.Get("traffic-pattern"), simpleQuery: q,
	}
	if strings.Contains(ports[0], "-") {
		m.Port, m.PortRange = 0, ports[0]
	} else {
		m.Port, _ = strconv.Atoi(ports[0])
	}
	if err := m.validate(); err != nil {
		return Mieru{}, err
	}
	for i := 1; i < len(ports); i++ {
		if err := m.addBinding(ports[i], transports[i]); err != nil {
			return Mieru{}, err
		}
	}
	return m, nil
}

func (m *Mieru) addBinding(port, transport string) error {
	if !validMieruBinding(port, transport) {
		return errors.New("无效的 Mieru 端口或传输协议")
	}
	m.otherBindings = append(m.otherBindings, MieruBinding{Port: port, Transport: transport})
	return nil
}

func (m Mieru) validate() error {
	if m.Name == "" || m.Server == "" || m.Username == "" {
		return errors.New("Mieru 配置缺少名称、服务器或用户名")
	}
	if m.Password == "" && (m.standardConfig == nil ||
		m.standardConfig.GetProfiles()[m.standardProfile].GetUser().GetHashedPassword() == "") {
		return errors.New("Mieru 配置缺少密码")
	}
	port := m.portValue()
	if strings.Contains(port, "-") != (m.PortRange != "") || !validMieruBinding(port, m.Transport) {
		return errors.New("无效的 Mieru 端口或传输协议")
	}
	if m.MTU != 0 && (m.MTU < 1280 || m.MTU > 1500) {
		return errors.New("无效的 Mieru MTU")
	}
	if _, ok := pb.MultiplexingLevel_value[m.Multiplexing]; m.Multiplexing != "" && !ok {
		return errors.New("无效的 Mieru 多路复用模式")
	}
	if _, ok := pb.HandshakeMode_value[m.HandshakeMode]; m.HandshakeMode != "" && !ok {
		return errors.New("无效的 Mieru 握手模式")
	}
	if m.TrafficPattern != "" {
		if _, err := trafficpattern.Decode(m.TrafficPattern); err != nil {
			return errors.New("无效的 Mieru 流量模式")
		}
	}
	return nil
}

func validMieruBinding(port, transport string) bool {
	first, last, ranged := strings.Cut(port, "-")
	start, err := strconv.Atoi(first)
	if err != nil || start < 1 || start > 65535 || (transport != "TCP" && transport != "UDP") {
		return false
	}
	if !ranged {
		return true
	}
	end, err := strconv.Atoi(last)
	return err == nil && end >= start && end <= 65535
}

func EncodeMieruURL(m Mieru) string {
	m.Server = strings.Trim(strings.TrimSpace(m.Server), "[]")
	m.PortRange = strings.TrimSpace(m.PortRange)
	m.Transport = strings.ToUpper(strings.TrimSpace(m.Transport))
	if m.Transport == "" {
		m.Transport = "TCP"
	}
	if m.Name == "" {
		m.Name = "default"
	}
	if m.validate() != nil {
		return ""
	}
	if m.standardConfig != nil {
		return encodeMieruStandard(m)
	}
	return encodeMieruSimple(m)
}

func encodeMieruSimple(m Mieru) string {
	q := maps.Clone(m.simpleQuery)
	if q == nil {
		q = url.Values{}
	}
	q.Set("profile", m.Name)
	q.Set("port", m.portValue())
	q.Set("protocol", m.Transport)
	for _, binding := range m.otherBindings {
		q.Add("port", binding.Port)
		q.Add("protocol", binding.Transport)
	}
	for key, value := range map[string]string{
		"multiplexing": m.Multiplexing, "handshake-mode": m.HandshakeMode, "traffic-pattern": m.TrafficPattern,
	} {
		q.Del(key)
		if value != "" {
			q.Set(key, value)
		}
	}
	q.Del("mtu")
	if m.MTU != 0 {
		q.Set("mtu", strconv.Itoa(m.MTU))
	}
	u := url.URL{Scheme: "mierus", Host: utils.WrapIPv6Host(m.Server), User: url.UserPassword(m.Username, m.Password), RawQuery: q.Encode()}
	return u.String()
}

func encodeMieruStandard(m Mieru) string {
	config := proto.CloneOf(m.standardConfig)
	profile := config.GetProfiles()[m.standardProfile]
	server := profile.GetServers()[0]
	binding := server.GetPortBindings()[0]

	profile.ProfileName, config.ActiveProfile = proto.String(m.Name), proto.String(m.Name)
	if profile.User.GetName() != m.Username || profile.User.GetPassword() != m.Password {
		profile.User.HashedPassword = nil
	}
	if m.Password == "" && profile.User.GetHashedPassword() == "" {
		return ""
	}
	profile.User.Name, profile.User.Password = proto.String(m.Username), proto.String(m.Password)
	if m.Server != server.GetDomainName() && (server.GetDomainName() != "" || m.Server != server.GetIpAddress()) {
		server.IpAddress, server.DomainName = nil, nil
		if net.ParseIP(m.Server) == nil {
			server.DomainName = proto.String(m.Server)
		} else {
			server.IpAddress = proto.String(m.Server)
		}
	}
	binding.Port, binding.PortRange = nil, nil
	if m.PortRange != "" {
		binding.PortRange = proto.String(m.PortRange)
	} else {
		port, _ := strconv.ParseInt(m.portValue(), 10, 32)
		binding.Port = proto.Int32(int32(port))
	}
	binding.Protocol = pb.TransportProtocol(pb.TransportProtocol_value[m.Transport]).Enum()
	profile.Mtu, profile.Multiplexing, profile.HandshakeMode, profile.TrafficPattern = nil, nil, nil, nil
	if m.MTU >= 1280 && m.MTU <= 1500 {
		profile.Mtu = proto.Int32(int32(m.MTU))
	}
	if m.Multiplexing != "" {
		profile.Multiplexing = &pb.MultiplexingConfig{Level: pb.MultiplexingLevel(pb.MultiplexingLevel_value[m.Multiplexing]).Enum()}
	}
	if m.HandshakeMode != "" {
		profile.HandshakeMode = pb.HandshakeMode(pb.HandshakeMode_value[m.HandshakeMode]).Enum()
	}
	if m.TrafficPattern != "" {
		profile.TrafficPattern, _ = trafficpattern.Decode(m.TrafficPattern)
	}
	data, err := proto.Marshal(config)
	if err != nil {
		return ""
	}
	return "mieru://" + base64.StdEncoding.EncodeToString(data)
}

func ConvertProxyToMieru(proxy Proxy) Mieru {
	return Mieru{
		Name: proxy.Name, Server: proxy.Server, Port: int(proxy.Port), PortRange: proxy.PortRange, Transport: proxy.Transport,
		Username: proxy.Username, Password: proxy.Password, Multiplexing: proxy.Multiplexing,
		TrafficPattern: proxy.TrafficPattern, HandshakeMode: proxy.HandshakeMode, MTU: proxy.MieruMTU, otherBindings: proxy.MieruBindings,
	}
}

func buildMieruProxy(link Urls, _ OutputConfig) (Proxy, error) {
	m, err := DecodeMieruURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	if m.Password == "" {
		return Proxy{}, errors.New("导出到 Clash/mihomo 需要 Mieru 明文密码")
	}
	if m.standardConfig != nil && len(m.standardConfig.GetProfiles()[m.standardProfile].GetServers()) > 1 {
		return Proxy{}, errors.New("导出到 Clash/mihomo 仅支持单服务器 Mieru 节点")
	}
	return Proxy{
		Name: m.Name, Type: "mieru", Server: m.Server, Port: FlexPort(utils.GetPortInt(m.Port)), PortRange: m.PortRange,
		Transport: m.Transport, Username: m.Username, Password: m.Password, Multiplexing: m.Multiplexing,
		TrafficPattern: m.TrafficPattern, HandshakeMode: m.HandshakeMode, MieruMTU: m.MTU,
		MieruBindings: m.otherBindings, Dialer_proxy: link.DialerProxyName,
	}, nil
}

// clashPortRange merges only bindings that describe one uninterrupted range
// with a single transport, as required by the Mihomo Mieru outbound.
func (p Proxy) clashPortRange() (Proxy, bool) {
	if len(p.MieruBindings) == 0 {
		return p, true
	}

	type interval struct{ start, end int }
	intervals := make([]interval, 0, len(p.MieruBindings)+1)
	port := p.PortRange
	if port == "" {
		port = strconv.Itoa(p.Port.Int())
	}
	bindings := append([]MieruBinding{{Port: port, Transport: p.Transport}}, p.MieruBindings...)
	for _, binding := range bindings {
		if binding.Transport != p.Transport {
			return Proxy{}, false
		}
		first, last, ranged := strings.Cut(binding.Port, "-")
		start, _ := strconv.Atoi(first)
		end := start
		if ranged {
			end, _ = strconv.Atoi(last)
		}
		intervals = append(intervals, interval{start, end})
	}
	slices.SortFunc(intervals, func(a, b interval) int { return a.start - b.start })

	start, end := intervals[0].start, intervals[0].end
	for _, next := range intervals[1:] {
		if next.start > end+1 {
			return Proxy{}, false
		}
		end = max(end, next.end)
	}
	p.Port, p.PortRange = FlexPort(start), ""
	if start != end {
		p.Port, p.PortRange = 0, fmt.Sprintf("%d-%d", start, end)
	}
	return p, true
}

// UpgradeLegacyMieruLink is only used by the startup migration.
func UpgradeLegacyMieruLink(link, extra string) (string, string, error) {
	u, err := url.Parse(strings.TrimSpace(link))
	if err != nil || u.Scheme != "mieru" || u.User == nil || u.Hostname() == "" {
		return "", extra, errors.New("无效的旧版 Mieru 链接")
	}
	q := u.Query()
	password, _ := u.User.Password()
	m := Mieru{
		Name: u.Fragment, Server: u.Hostname(), Port: u.Port(), PortRange: q.Get("portRange"),
		Transport: q.Get("transport"), Username: u.User.Username(), Password: password,
		Multiplexing: q.Get("multiplexing"), HandshakeMode: q.Get("handshake-mode"), TrafficPattern: q.Get("trafficPattern"),
	}
	if m.PortRange == "" {
		m.PortRange = q.Get("port-range")
	}
	if m.TrafficPattern == "" {
		m.TrafficPattern = q.Get("traffic-pattern")
	}
	if u.Port() == "" && q.Has("port") && q.Has("protocol") {
		u.Scheme = "mierus"
		m, err = decodeMieruSimple(u.String())
		if err != nil {
			return "", extra, err
		}
	}
	if m.Name == "" {
		m.Name = formatURLHostPort(m.Server, m.portValue())
	}

	// Older YAML imports retained the handshake mode as an unmodeled field.
	var proxy Proxy
	if err := ApplyClashExtra(&proxy, extra); err != nil {
		return "", extra, err
	}
	if mapping := yamlMappingNode(&proxy.ClashExtra); mapping != nil {
		for i := 0; i+1 < len(mapping.Content); i += 2 {
			if mapping.Content[i].Value == "handshake-mode" {
				if m.HandshakeMode == "" {
					m.HandshakeMode = mapping.Content[i+1].Value
				}
				mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
				extra, err = EncodeClashExtra(proxy)
				if err != nil {
					return "", extra, err
				}
				break
			}
		}
	}
	canonical := EncodeMieruURL(m)
	if canonical == "" {
		return "", extra, errors.New("无法转换旧版 Mieru 链接")
	}
	return canonical, extra, nil
}
