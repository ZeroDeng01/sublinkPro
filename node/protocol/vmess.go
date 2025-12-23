package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
	"sublink/utils"
)

type Vmess struct {
	Add  string      `json:"add,omitempty"` // 服务器地址
	Aid  interface{} `json:"aid,omitempty"`
	Alpn string      `json:"alpn,omitempty"`
	Fp   string      `json:"fp,omitempty"`
	Host string      `json:"host,omitempty"`
	Id   string      `json:"id,omitempty"`
	Net  string      `json:"net,omitempty"`
	Path string      `json:"path,omitempty"`
	Port interface{} `json:"port,omitempty"`
	Ps   string      `json:"ps,omitempty"`
	Scy  string      `json:"scy,omitempty"`
	Sni  string      `json:"sni,omitempty"`
	Tls  string      `json:"tls,omitempty"`
	Type string      `json:"type,omitempty"`
	V    string      `json:"v,omitempty"`
}

// 开发者测试
func CallVmessURL() {
	vmess := Vmess{
		Add:  "xx.xxx.ru",
		Port: "2095",
		Aid:  0,
		Scy:  "auto",
		Net:  "ws",
		Type: "none",
		Id:   "7a737f41-b792-4260-94ff-3d864da67380",
		Host: "xx.xxx.ru",
		Path: "/",
		Tls:  "",
	}
	fmt.Println(EncodeVmessURL(vmess))
}

// vmess 编码
func EncodeVmessURL(v Vmess) string {
	// 如果备注为空，则使用服务器地址+端口
	if v.Ps == "" {
		v.Ps = v.Add + ":" + utils.GetPortString(v.Port)
	}
	// 如果版本为空，则默认为2
	if v.V == "" {
		v.V = "2"
	}
	param, _ := json.Marshal(v)
	return "vmess://" + utils.Base64Encode(string(param))
}

// vmess 解码
func DecodeVMESSURL(s string) (Vmess, error) {
	if !strings.Contains(s, "vmess://") {
		return Vmess{}, fmt.Errorf("非vmess协议:%s", s)
	}
	param := strings.Split(s, "://")[1]
	param = utils.Base64Decode(strings.TrimSpace(param))
	// fmt.Println(param)
	var vmess Vmess
	err := json.Unmarshal([]byte(param), &vmess)
	if err != nil {
		utils.Error("❌节点解析错误：%v  【节点：%s】", err, param)
		return Vmess{}, fmt.Errorf("json格式化失败:%s", param)
	}
	if !utils.IsUUID(vmess.Id) {
		utils.Error("❌节点解析错误：%v  【节点：%s】", "UUID格式错误", param)
		return Vmess{}, fmt.Errorf("uuid格式错误:%s", vmess.Id)
	}
	vmess.Add = utils.UnwrapIPv6Host(vmess.Add)
	if vmess.Scy == "" {
		vmess.Scy = "auto"
	}
	// 如果备注为空，则使用服务器地址+端口
	if vmess.Ps == "" {
		vmess.Ps = vmess.Add + ":" + utils.GetPortString(vmess.Port)
	}
	if utils.CheckEnvironment() {
		fmt.Println("服务器地址", vmess.Add)
		fmt.Println("端口", vmess.Port)
		fmt.Println("path", vmess.Path)
		fmt.Println("uuid", vmess.Id)
		fmt.Println("alterId", vmess.Aid)
		fmt.Println("cipher", vmess.Scy)
		fmt.Println("client-fingerprint", vmess.Fp)
		fmt.Println("network", vmess.Net)
		fmt.Println("tls", vmess.Tls)
		fmt.Println("备注", vmess.Ps)
	}
	return vmess, nil
}
