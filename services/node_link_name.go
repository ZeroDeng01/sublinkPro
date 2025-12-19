package services

import (
	"fmt"
	"strings"
	"sublink/database"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/utils"

	"gorm.io/gorm"
)

// UpdateNodeLinkName 修改节点的原始名称并重新编码Link
// 流程：
// 1. 根据 Link 前缀判断协议类型
// 2. 调用对应协议的 Decode 函数解析
// 3. 修改协议对象中的名称字段
// 4. 调用对应协议的 Encode 函数重新生成 Link
// 5. 更新数据库中的 Link 和 LinkName
func UpdateNodeLinkName(nodeID int, newLinkName string) error {
	// 获取节点
	var node models.Node
	node.ID = nodeID
	if err := node.GetByID(); err != nil {
		return fmt.Errorf("获取节点失败: %w", err)
	}

	if node.LinkName == node.Name {
		node.Name = newLinkName
	}

	// 解析协议类型并处理
	newLink, err := updateLinkWithNewName(node.Link, newLinkName)
	if err != nil {
		return fmt.Errorf("更新节点名称失败: %w", err)
	}

	// 检查新 Link 是否与其他节点冲突（排除当前节点）
	var existingNode models.Node
	err = database.DB.Where("link = ? ", newLink).First(&existingNode).Error
	if err == nil {
		return fmt.Errorf("已存在相同连接的节点: %s", existingNode.Name)
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("检查节点冲突失败: %w", err)
	}

	// 更新数据库
	err = database.DB.Model(&models.Node{}).Where("id = ?", nodeID).Updates(map[string]interface{}{
		"link":      newLink,
		"link_name": newLinkName,
		"name":      node.Name,
	}).Error
	if err != nil {
		return fmt.Errorf("更新数据库失败: %w", err)
	}

	// 更新缓存
	node.Link = newLink
	node.LinkName = newLinkName
	models.UpdateNodeCache(nodeID, node)

	utils.Info("节点 [%s] 原始名称从 [%s] 修改为 [%s]", node.Name, node.LinkName, newLinkName)
	return nil
}

// updateLinkWithNewName 根据协议类型解码链接，修改名称后重新编码
func updateLinkWithNewName(link string, newName string) (string, error) {
	linkLower := strings.ToLower(link)

	switch {
	case strings.HasPrefix(linkLower, "vmess://"):
		return updateVmessName(link, newName)
	case strings.HasPrefix(linkLower, "vless://"):
		return updateVlessName(link, newName)
	case strings.HasPrefix(linkLower, "trojan://"):
		return updateTrojanName(link, newName)
	case strings.HasPrefix(linkLower, "ss://"):
		return updateSSName(link, newName)
	case strings.HasPrefix(linkLower, "ssr://"):
		return updateSSRName(link, newName)
	case strings.HasPrefix(linkLower, "hysteria://") || strings.HasPrefix(linkLower, "hy://"):
		return updateHysteriaName(link, newName)
	case strings.HasPrefix(linkLower, "hysteria2://") || strings.HasPrefix(linkLower, "hy2://"):
		return updateHysteria2Name(link, newName)
	case strings.HasPrefix(linkLower, "tuic://"):
		return updateTuicName(link, newName)
	case strings.HasPrefix(linkLower, "socks5://"):
		return updateSocks5Name(link, newName)
	case strings.HasPrefix(linkLower, "anytls://"):
		return updateAnyTLSName(link, newName)
	default:
		return "", fmt.Errorf("不支持的协议类型")
	}
}

// 以下是各协议的名称更新函数

func updateVmessName(link string, newName string) (string, error) {
	vmess, err := protocol.DecodeVMESSURL(link)
	if err != nil {
		return "", err
	}
	vmess.Ps = newName
	return protocol.EncodeVmessURL(vmess), nil
}

func updateVlessName(link string, newName string) (string, error) {
	vless, err := protocol.DecodeVLESSURL(link)
	if err != nil {
		return "", err
	}
	vless.Name = newName
	return protocol.EncodeVLESSURL(vless), nil
}

func updateTrojanName(link string, newName string) (string, error) {
	trojan, err := protocol.DecodeTrojanURL(link)
	if err != nil {
		return "", err
	}
	trojan.Name = newName
	return protocol.EncodeTrojanURL(trojan), nil
}

func updateSSName(link string, newName string) (string, error) {
	ss, err := protocol.DecodeSSURL(link)
	if err != nil {
		return "", err
	}
	ss.Name = newName
	return protocol.EncodeSSURL(ss), nil
}

func updateSSRName(link string, newName string) (string, error) {
	ssr, err := protocol.DecodeSSRURL(link)
	if err != nil {
		return "", err
	}
	ssr.Qurey.Remarks = newName
	return protocol.EncodeSSRURL(ssr), nil
}

func updateHysteriaName(link string, newName string) (string, error) {
	hy, err := protocol.DecodeHYURL(link)
	if err != nil {
		return "", err
	}
	hy.Name = newName
	return protocol.EncodeHYURL(hy), nil
}

func updateHysteria2Name(link string, newName string) (string, error) {
	hy2, err := protocol.DecodeHY2URL(link)
	if err != nil {
		return "", err
	}
	hy2.Name = newName
	return protocol.EncodeHY2URL(hy2), nil
}

func updateTuicName(link string, newName string) (string, error) {
	tuic, err := protocol.DecodeTuicURL(link)
	if err != nil {
		return "", err
	}
	tuic.Name = newName
	return protocol.EncodeTuicURL(tuic), nil
}

func updateSocks5Name(link string, newName string) (string, error) {
	socks5, err := protocol.DecodeSocks5URL(link)
	if err != nil {
		return "", err
	}
	socks5.Name = newName
	return protocol.EncodeSocks5URL(socks5), nil
}

func updateAnyTLSName(link string, newName string) (string, error) {
	anytls, err := protocol.DecodeAnyTLSURL(link)
	if err != nil {
		return "", err
	}
	anytls.Name = newName
	return protocol.EncodeAnyTLSURL(anytls), nil
}
