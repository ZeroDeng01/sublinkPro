package protocol

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

// ss匹配规则
type Ss struct {
	Param  Param
	Server string
	Port   interface{}
	Name   string
	Type   string
}
type Param struct {
	Cipher   string
	Password string
}

func parsingSS(s string) (string, string, string) {
	/* ss url编码分为三部分：加密方式、服务器地址和端口、备注
	://和@之前为第一部分 @到#之间为第二部分 #之后为第三部分
	第一部分 为加密方式和密码，格式为：加密方式:密码	示例：aes-128-gcm:123456
	第二部分 为服务器地址和端口，格式为：服务器地址:端口	示例：xxx.xxx:12345
	第三部分 为备注，格式为：#备注	示例：#备注
	*/
	u, err := url.Parse(s)
	if err != nil {
		log.Println("ss url parse fail.", err)
		return "", "", ""
	}
	if u.Scheme != "ss" {
		log.Println("ss url parse fail, not ss url.")
		return "", "", ""
	}
	// 处理url全编码的情况
	if u.User == nil {
		// 截取ss://后的字符串
		raw := s[5:]
		s = "ss://" + utils.Base64Decode(raw)
		u, err = url.Parse(s)
	}
	var auth, addr, name string
	auth = u.User.String()
	if u.Host != "" {
		addr = u.Host
	}
	if u.Fragment != "" {
		name = u.Fragment
	}
	return auth, addr, name
}

// 开发者测试
func CallSSURL() {
	ss := Ss{}
	// ss.Name = "测试"
	ss.Server = "baidu.com"
	ss.Port = 443
	ss.Param.Cipher = "2022-blake3-aes-256-gcm"
	ss.Param.Password = "asdasd"
	fmt.Println(EncodeSSURL(ss))
}

// ss 编码输出
func EncodeSSURL(s Ss) string {
	//编码格式 ss://base64(base64(method:password)@hostname:port)
	p := utils.Base64Encode(s.Param.Cipher + ":" + s.Param.Password)
	// 假设备注没有使用服务器加端口命名
	if s.Name == "" {
		s.Name = s.Server + ":" + utils.GetPortString(s.Port)
	}
	param := fmt.Sprintf("%s@%s:%s#%s",
		p,
		s.Server,
		utils.GetPortString(s.Port),
		s.Name,
	)
	return "ss://" + param
}

func DecodeSSURL(s string) (Ss, error) {
	// 解析ss链接
	param, addr, name := parsingSS(s)
	// base64解码
	param = utils.Base64Decode(param)
	// 判断是否为空
	if param == "" || addr == "" {
		return Ss{}, fmt.Errorf("invalid SS URL")
	}
	// 解析参数
	parts := strings.Split(addr, ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])
	server := strings.Replace(utils.UnwrapIPv6Host(addr), ":"+parts[len(parts)-1], "", -1)
	cipher := strings.Split(param, ":")[0]
	password := strings.Replace(param, cipher+":", "", 1)
	// 如果没有备注则使用服务器加端口命名
	if name == "" {
		name = addr
	}
	// 开发环境输出结果
	if utils.CheckEnvironment() {
		fmt.Println("Param:", utils.Base64Decode(param))
		fmt.Println("Server", server)
		fmt.Println("Port", port)
		fmt.Println("Name:", name)
		fmt.Println("Cipher:", cipher)
		fmt.Println("Password:", password)
	}
	// 返回结果
	return Ss{
		Param: Param{
			Cipher:   cipher,
			Password: password,
		},
		Server: server,
		Port:   port,
		Name:   name,
		Type:   "ss",
	}, nil
}
