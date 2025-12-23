package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sublink/utils"
)

func CallSSRURL() {
	ssr := new(Ssr)
	ssr.Server = "xx.com"
	ssr.Port = 443
	ssr.Protocol = "auth_aes128_md5"
	ssr.Method = "aes-256-cfb"
	ssr.Obfs = "tls1.2_ticket_auth"
	ssr.Password = "123456"
	ssr.Qurey = Ssrquery{
		Obfsparam: "",
		Remarks:   "没有名字",
	}
	cc := EncodeSSRURL(*ssr)
	fmt.Println(cc)
}

// ssr格式编码输出
func EncodeSSRURL(s Ssr) string {
	/*编码格式
	ssr://base64(host:port:protocol:method:obfs:base64(password)/?obfsparam=base64(obfsparam)&protoparam=base64(protoparam)&remarks=base64(remarks)&group=base64(group))
	*/
	obfsparam := "obfsparam=" + utils.Base64Encode(s.Qurey.Obfsparam)
	remarks := "remarks=" + utils.Base64Encode(s.Qurey.Remarks)
	// 如果没有备注默认使用服务器+端口作为备注
	if s.Qurey.Remarks == "" {
		server_port := utils.Base64Encode(s.Server + ":" + utils.GetPortString(s.Port))
		remarks = fmt.Sprintf("remarks=%s", server_port)
	}
	param := fmt.Sprintf("%s:%s:%s:%s:%s:%s/?%s&%s",
		s.Server,
		utils.GetPortString(s.Port),
		s.Protocol,
		s.Method,
		s.Obfs,
		utils.Base64Encode(s.Password),
		obfsparam,
		remarks,
	)
	return "ssr://" + utils.Base64Encode(param)

}

// ssr解码
func DecodeSSRURL(s string) (Ssr, error) {
	/*解析格式
	ssr://base64(host:port:protocol:method:obfs:base64(password)/?obfsparam=base64(obfsparam)&protoparam=base64(protoparam)&remarks=base64(remarks)&group=base64(group))
	*/
	// 处理url链接中的base64编码
	parts := strings.SplitN(s, "ssr://", 2)
	if len(parts) != 2 {
		return Ssr{}, errors.New("invalid SSR URL")
	}
	s = parts[0] + utils.Base64Decode(parts[1])
	// 检查是否包含"/?" 如果有就是有备注信息
	var remarks, obfsparam string
	if strings.Contains(s, "/?") {
		// 解析备注信息
		query := strings.Split(s, "/?")[1]
		s = strings.Replace(s, "/?"+query, "", 1)
		paramMap := make(map[string]string)
		if strings.Contains(query, "&") {
			params := strings.Split(query, "&")
			for _, param := range params {
				parts := strings.SplitN(param, "=", 2)
				if len(parts) != 2 {
					fmt.Println("Invalid parameter: ", param)
					continue
				}
				paramMap[parts[0]] = parts[1]
			}
		} else {
			q := strings.Split(query, "=")
			paramMap[q[0]] = q[1]
		}
		remarks = utils.Base64Decode(paramMap["remarks"])
		obfsparam = utils.Base64Decode(paramMap["obfsparam"])
		defer func() {
			if utils.CheckEnvironment() {
				fmt.Println("remarks", remarks)
				fmt.Println("obfsparam", obfsparam)
			}
		}()
	}
	// 反着解析参数 怕有ipv6地址冒号混淆
	param := strings.Split(s, ":")
	if len(param) < 6 {
		return Ssr{}, errors.New("长度没有6")
	}
	password := param[len(param)-1]
	obfs := param[len(param)-2]
	method := param[len(param)-3]
	protocol := param[len(param)-4]
	port, _ := strconv.Atoi(param[len(param)-5])
	server := utils.UnwrapIPv6Host(param[len(param)-6])
	// 如果没有备注默认使用服务器+端口作为备注
	if remarks == "" {
		remarks = server + ":" + strconv.Itoa(port)
	}
	if utils.CheckEnvironment() {
		fmt.Println("password", password)
		fmt.Println("obfs", obfs)
		fmt.Println("method", method)
		fmt.Println("protocol", protocol)
		fmt.Println("port", port)
		fmt.Println("server", server)
	}
	return Ssr{
		Server:   server,
		Port:     port,
		Protocol: protocol,
		Method:   method,
		Obfs:     obfs,
		Password: password,
		Qurey: Ssrquery{
			Obfsparam: obfsparam,
			Remarks:   remarks,
		},
		Type: "ssr",
	}, nil
}

type Ssr struct {
	Server   string
	Port     interface{}
	Protocol string
	Method   string
	Obfs     string
	Password string
	Qurey    Ssrquery
	Type     string
}
type Ssrquery struct {
	Obfsparam string
	Remarks   string
}
