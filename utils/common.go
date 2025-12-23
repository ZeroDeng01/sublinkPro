package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// 检查环境
func CheckEnvironment() bool {
	APP_ENV := os.Getenv("APP_ENV")
	if APP_ENV == "" {
		return false
	}
	if strings.Contains(APP_ENV, "development") {
		return true
	}
	return false
}

// Base62字符集
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// ToBase62 将字节数组转换为Base62字符串
func ToBase62(data []byte) string {
	bi := new(big.Int).SetBytes(data)
	result := ""

	for bi.Sign() > 0 {
		mod := new(big.Int)
		bi.DivMod(bi, big.NewInt(62), mod)
		result = string(base62Chars[mod.Int64()]) + result
	}

	// 确保至少包含一个字符
	if result == "" {
		result = "0"
	}

	return result
}

// FromBase62 将Base62字符串转换回字节数组
func FromBase62(s string) ([]byte, error) {
	bi := new(big.Int)

	for _, c := range s {
		bi.Mul(bi, big.NewInt(62))
		pos := strings.IndexRune(base62Chars, c)
		if pos < 0 {
			return nil, fmt.Errorf("无效字符: %c", c)
		}
		bi.Add(bi, big.NewInt(int64(pos)))
	}

	return bi.Bytes(), nil
}

// EncryptUserIDCompact 使用更紧凑的方法加密用户ID (Base62版本)
func EncryptUserIDCompact(userID int, key []byte) (string, error) {

	// 简单加密方法：将userID与唯一的密钥混合
	// 使用SHA256生成派生密钥
	h := sha256.New()
	h.Write(key)

	// 将userID转换为字节并与派生密钥混合
	userIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(userIDBytes, uint32(userID))

	// 将userID与派生密钥的前4字节进行XOR操作
	derivedKey := h.Sum(nil)[:4]
	encrypted := make([]byte, 4)
	for i := 0; i < 4; i++ {
		encrypted[i] = userIDBytes[i] ^ derivedKey[i]
	}

	// 转换为Base62编码字符串，使结果更短
	return ToBase62(encrypted), nil
}

// DecryptUserIDCompact 解密用户ID (Base62版本)
func DecryptUserIDCompact(encrypted string, key []byte) (int, error) {
	// 使用SHA256生成相同的密钥派生密钥
	h := sha256.New()
	h.Write(key)
	derivedKey := h.Sum(nil)[:4]

	// Base62解码
	encryptedBytes, err := FromBase62(encrypted)
	if err != nil {
		Error("Base62解码失败: %v", err)
		return 0, fmt.Errorf("Base62解码失败: %w", err)
	}

	// 确保长度不超过4字节
	if len(encryptedBytes) > 4 {
		Error("解码后数据长度超过4字节: %d", len(encryptedBytes))
		return 0, fmt.Errorf("解码后数据长度超过4字节: %d", len(encryptedBytes))
	}

	// 如果长度不足4字节，则在前面填充0
	if len(encryptedBytes) < 4 {
		padding := make([]byte, 4-len(encryptedBytes))
		encryptedBytes = append(padding, encryptedBytes...)
	}

	// 与加密使用相同的XOR操作解密
	decrypted := make([]byte, 4)
	for i := 0; i < 4; i++ {
		decrypted[i] = encryptedBytes[i] ^ derivedKey[i]
	}

	// 转换回用户ID，使用uint32确保不会出现负数
	userID := int(binary.BigEndian.Uint32(decrypted))
	return userID, nil
}

// ipv6地址匹配规则
func UnwrapIPv6Host(s string) string {
	pattern := `\[([0-9a-fA-F:]+)\]`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(s)
	if len(match) > 0 {
		return match[1]
	} else {
		return s
	}
}

func WrapIPv6Host(s string) string {
	// 判断是否已经是 [IPv6] 格式
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return s
	}

	// 判断是否是 IPv6 地址（包含冒号但不是域名或 IPv4）
	if strings.Contains(s, ":") && !strings.ContainsAny(s, ".[]") {
		return "[" + s + "]"
	}
	return s
}

// 判断是否需要补全
func IsBase64makeup(s string) string {
	l := len(s)
	if l%4 != 0 {
		return s + strings.Repeat("=", 4-l%4)
	}
	return s
}

// base64编码
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// base64解码
func Base64Decode(s string) string {
	// 去除空格
	s = strings.ReplaceAll(s, " ", "")
	// 判断是否有特殊字符来判断是标准base64还是url base64
	match, err := regexp.MatchString(`[_-]`, s)
	if err != nil {
		fmt.Println(err)
	}
	if !match {
		// 默认使用标准解码
		encoded := IsBase64makeup(s)
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return s // 返回原字符串
		}
		decoded_str := string(decoded)
		return decoded_str

	} else {
		// 如果有特殊字符则使用URL解码
		encoded := IsBase64makeup(s)
		decoded, err := base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			return s // 返回原字符串
		}
		decoded_str := string(decoded)
		return decoded_str
	}
}

// base64解码不自动补齐
func Base64Decode2(s string) string {
	// 去除空格
	s = strings.ReplaceAll(s, " ", "")
	// 判断是否有特殊字符来判断是标准base64还是url base64
	match, err := regexp.MatchString(`[_-]`, s)
	if err != nil {
		fmt.Println(err)
	}
	if !match {
		// 默认使用标准解码
		decoded, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return s // 返回原字符串
		}
		decoded_str := string(decoded)
		return decoded_str

	} else {
		// 如果有特殊字符则使用URL解码
		decoded, err := base64.URLEncoding.DecodeString(s)
		if err != nil {
			return s // 返回原字符串
		}
		decoded_str := string(decoded)
		return decoded_str
	}
}

// RandString 生成随机字符串
func RandString(number int) string {
	str := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	// 用 []byte 直接构造字符串
	n := rand.Intn(number) + 1 // 防止生成空字符串，范围是1到31
	randomString := make([]byte, n)
	for i := 0; i < n; i++ {
		randomIndex := rand.Intn(len(str))
		randomString[i] = str[randomIndex]
	}
	Secret := string(randomString)
	return Secret
}

// IsIpInCidr 判断IP是否在CIDR范围内
func IsIpInCidr(cIP string, sIP string) bool {
	ips := strings.Split(sIP, ",")
	for _, ip := range ips {
		if strings.Contains(ip, "/") {
			_, ipNet, err := net.ParseCIDR(ip)
			if err != nil {
				return false
			}
			return ipNet.Contains(net.ParseIP(cIP))
		} else {
			return ip == cIP
		}
	}
	return false
}

// IpFormatValidation IP格式检测
func IpFormatValidation(ipString string) bool {
	ips := strings.Split(ipString, ",")
	for _, ip := range ips {
		_, _, err := net.ParseCIDR(ip)
		if err != nil {
			parsedIP := net.ParseIP(ip)
			if parsedIP == nil {
				return false
			}
		}
	}
	return true
}

// FormatBytes 格式化字节数
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsUUID 检测是否为UUID
func IsUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// GetPortString 将 interface{} 类型的端口转换为字符串
func GetPortString(port interface{}) string {
	switch p := port.(type) {
	case int:
		return strconv.Itoa(p)
	case float64:
		return strconv.Itoa(int(p))
	case string:
		return p
	default:
		return fmt.Sprintf("%v", port)
	}
}

// GetPortInt 将 interface{} 类型的端口转换为整数
func GetPortInt(port interface{}) int {
	switch p := port.(type) {
	case int:
		return p
	case float64:
		return int(p)
	case string:
		val, _ := strconv.Atoi(p)
		return val
	default:
		return 0
	}
}
