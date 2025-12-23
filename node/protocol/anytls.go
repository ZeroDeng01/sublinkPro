package protocol

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

type AnyTLS struct {
	Name              string
	Server            string
	Port              interface{}
	Password          string
	SkipCertVerify    bool
	SNI               string
	ClientFingerprint string
}

func DecodeAnyTLSURL(s string) (AnyTLS, error) {

	if !strings.Contains(s, "anytls://") {
		return AnyTLS{}, fmt.Errorf("非anytls协议: %s", s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return AnyTLS{}, fmt.Errorf("url parse error: %v", err)
	}
	var anyTLS AnyTLS
	name := u.Fragment
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		fmt.Println("AnyTLS SplitHostPort error", err)
		return AnyTLS{}, err
	}
	anyTLS.Server = host
	rawPort := port
	if rawPort == "" {
		rawPort = "443"
	}
	anyTLS.Port, err = strconv.Atoi(rawPort)
	if err != nil {
		fmt.Println("AnyTLS Port conversion failed:", err)
		return AnyTLS{}, err
	}
	anyTLS.Password = u.User.Username()
	skipCertVerify := u.Query().Get("insecure")
	if skipCertVerify != "" {
		anyTLS.SkipCertVerify, err = strconv.ParseBool(skipCertVerify)
	}
	if err != nil {
		fmt.Println("AnyTLS SkipCertVerify conversion failed:", err)
		return AnyTLS{}, err
	}
	anyTLS.SNI = u.Query().Get("sni")
	anyTLS.ClientFingerprint = u.Query().Get("fp")

	if name == "" {
		anyTLS.Name = u.Host
	} else {
		anyTLS.Name = name
	}
	return anyTLS, nil
}

// EncodeAnyTLSURL anytls 编码
func EncodeAnyTLSURL(a AnyTLS) string {
	u := url.URL{
		Scheme:   "anytls",
		User:     url.User(a.Password),
		Host:     fmt.Sprintf("%s:%s", a.Server, utils.GetPortString(a.Port)),
		Fragment: a.Name,
	}
	q := u.Query()
	if a.SkipCertVerify {
		q.Set("insecure", "1")
	}
	if a.SNI != "" {
		q.Set("sni", a.SNI)
	}
	if a.ClientFingerprint != "" {
		q.Set("fp", a.ClientFingerprint)
	}
	u.RawQuery = q.Encode()
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if a.Name == "" {
		u.Fragment = fmt.Sprintf("%s:%s", a.Server, utils.GetPortString(a.Port))
	}
	return u.String()
}
