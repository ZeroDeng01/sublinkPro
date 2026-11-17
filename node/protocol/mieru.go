package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("mieru", []string{"mieru://"}, "Mieru", "#5f6caf", "M", Mieru{}, "Name", DecodeMieruURL, EncodeMieruURL, func(m Mieru) LinkIdentity {
		port := utils.GetPortString(m.Port)
		if strings.TrimSpace(m.PortRange) != "" {
			port = strings.TrimSpace(m.PortRange)
		}
		return buildIdentity("mieru", m.Name, m.Server, port)
	},
		FieldMeta{Name: "Name", Label: "Node name", Type: "string", Group: "basic"},
		FieldMeta{Name: "Server", Label: "Server address", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "Port", Type: "int", Group: "basic"},
		FieldMeta{Name: "PortRange", Label: "Port range", Type: "string", Group: "basic", Placeholder: "2090-2099", Advanced: true},
		FieldMeta{Name: "Transport", Label: "Transport protocol", Type: "string", Group: "transport", Options: []string{"TCP", "UDP"}},
		FieldMeta{Name: "Username", Label: "Username", Type: "string", Group: "auth"},
		FieldMeta{Name: "Password", Label: "Password", Type: "string", Group: "auth", Secret: true},
		FieldMeta{Name: "Multiplexing", Label: "Multiplexing", Type: "string", Group: "transport", Options: []string{"MULTIPLEXING_OFF", "MULTIPLEXING_LOW", "MULTIPLEXING_MIDDLE", "MULTIPLEXING_HIGH"}, Advanced: true},
		FieldMeta{Name: "TrafficPattern", Label: "Traffic mode", Type: "string", Group: "transport", Advanced: true},
	).WithClientSupport(ClientClash, ClientMihomo).WithClientSupportAliases("mierus://")
	MustRegisterProtocol(newProxyProtocolSpec(base, func(link Urls, _ OutputConfig) (Proxy, error) {
		return buildMieruProxy(link)
	}, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "mieru")
	}, ConvertProxyToMieru, EncodeMieruURL))
}

// Mieru stores SublinkPro's internal editable Mieru URL shape.
// Mihomo documents Mieru as Clash YAML fields; SublinkPro uses this URL only
// for raw editing and import/export round-trips.
type Mieru struct {
	Name           string
	Server         string
	Port           any
	PortRange      string
	Transport      string
	Username       string
	Password       string
	Multiplexing   string
	TrafficPattern string
}

func DecodeMieruURL(s string) (Mieru, error) {
	u, err := url.Parse(s)
	if err != nil {
		return Mieru{}, fmt.Errorf("url parse error: %v", err)
	}
	if strings.ToLower(u.Scheme) != "mieru" {
		return Mieru{}, fmt.Errorf("not a mieru link: %s", s)
	}

	server := u.Hostname()
	if server == "" {
		return Mieru{}, fmt.Errorf("missing server address: %s", s)
	}

	query := u.Query()
	portRange := strings.TrimSpace(query.Get("portRange"))
	if portRange == "" {
		portRange = strings.TrimSpace(query.Get("port-range"))
	}

	rawPort := u.Port()
	port := 0
	if rawPort != "" {
		port, err = strconv.Atoi(rawPort)
		if err != nil {
			return Mieru{}, fmt.Errorf("mieru port conversion failed: %w", err)
		}
	}
	if portRange != "" {
		port = 0
	}
	if port == 0 && portRange == "" {
		return Mieru{}, fmt.Errorf("missing port or port range: %s", s)
	}

	password, _ := u.User.Password()
	trafficPattern := query.Get("trafficPattern")
	if trafficPattern == "" {
		trafficPattern = query.Get("traffic-pattern")
	}

	name := u.Fragment
	if name == "" {
		if portRange != "" {
			name = fmt.Sprintf("%s:%s", server, portRange)
		} else {
			name = fmt.Sprintf("%s:%d", server, port)
		}
	}

	return Mieru{
		Name:           name,
		Server:         server,
		Port:           port,
		PortRange:      portRange,
		Transport:      query.Get("transport"),
		Username:       u.User.Username(),
		Password:       password,
		Multiplexing:   query.Get("multiplexing"),
		TrafficPattern: trafficPattern,
	}, nil
}

func EncodeMieruURL(m Mieru) string {
	server := utils.WrapIPv6Host(strings.TrimSpace(m.Server))
	portRange := strings.TrimSpace(m.PortRange)
	host := server
	if portRange == "" {
		port := strings.TrimSpace(utils.GetPortString(m.Port))
		if port != "" && port != "0" && port != "<nil>" {
			host = formatURLHostPort(server, port)
		}
	}

	u := url.URL{
		Scheme:   "mieru",
		Host:     host,
		Fragment: m.Name,
	}
	if m.Username != "" {
		u.User = url.UserPassword(m.Username, m.Password)
	}

	q := u.Query()
	if portRange != "" {
		q.Set("portRange", portRange)
	}
	if m.Transport != "" {
		q.Set("transport", m.Transport)
	}
	if m.Multiplexing != "" {
		q.Set("multiplexing", m.Multiplexing)
	}
	if m.TrafficPattern != "" {
		q.Set("trafficPattern", m.TrafficPattern)
	}
	u.RawQuery = q.Encode()

	if u.Fragment == "" {
		if portRange != "" {
			u.Fragment = fmt.Sprintf("%s:%s", strings.Trim(server, "[]"), portRange)
		} else if port := strings.TrimSpace(utils.GetPortString(m.Port)); port != "" && port != "0" && port != "<nil>" {
			u.Fragment = fmt.Sprintf("%s:%s", strings.Trim(server, "[]"), utils.GetPortString(m.Port))
		} else {
			u.Fragment = strings.Trim(server, "[]")
		}
	}

	return u.String()
}

func ConvertProxyToMieru(proxy Proxy) Mieru {
	return Mieru{
		Name:           proxy.Name,
		Server:         proxy.Server,
		Port:           int(proxy.Port),
		PortRange:      proxy.PortRange,
		Transport:      proxy.Transport,
		Username:       proxy.Username,
		Password:       proxy.Password,
		Multiplexing:   proxy.Multiplexing,
		TrafficPattern: proxy.TrafficPattern,
	}
}

func buildMieruProxy(link Urls) (Proxy, error) {
	mieru, err := DecodeMieruURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}

	return Proxy{
		Name:           mieru.Name,
		Type:           "mieru",
		Server:         mieru.Server,
		Port:           FlexPort(utils.GetPortInt(mieru.Port)),
		PortRange:      mieru.PortRange,
		Transport:      mieru.Transport,
		Username:       mieru.Username,
		Password:       mieru.Password,
		Multiplexing:   mieru.Multiplexing,
		TrafficPattern: mieru.TrafficPattern,
		Dialer_proxy:   link.DialerProxyName,
	}, nil
}
