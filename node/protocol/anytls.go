package protocol

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type AnyTLS struct {
	Name              string
	Server            string
	Port              int
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
	anyTLS.Port, err = strconv.Atoi(port)
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
