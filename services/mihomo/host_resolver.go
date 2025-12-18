package mihomo

import (
	"net/netip"
	"sublink/models"
	"sublink/utils"

	"github.com/metacubex/mihomo/component/resolver"
	"github.com/metacubex/mihomo/component/trie"
)

// SyncHostsFromDB 从数据库同步Host配置到mihomo resolver
// 将项目Host模块的域名映射注入到resolver.DefaultHosts
// mihomo的DNS解析会优先查询DefaultHosts，命中则直接返回配置的IP
// 如果Host模块没有配置，则不做任何操作，resolver使用正常DNS解析
func SyncHostsFromDB() error {
	hosts := models.GetAllHosts()
	if len(hosts) == 0 {
		utils.Debug("无Host配置需要同步")
		return nil
	}

	// 创建新的DomainTrie并填充数据
	newHosts := trie.New[resolver.HostValue]()
	successCount := 0

	for _, h := range hosts {
		// 解析IP地址
		ip, err := netip.ParseAddr(h.IP)
		if err != nil {
			utils.Warn("解析Host IP失败: %s -> %s, 错误: %v", h.Hostname, h.IP, err)
			continue
		}

		// 创建HostValue（IP模式，非域名重定向）
		hostValue := resolver.HostValue{
			IsDomain: false,
			IPs:      []netip.Addr{ip.Unmap()},
		}

		// 插入到trie（支持通配符格式：*.example.com, .example.com）
		if err := newHosts.Insert(h.Hostname, hostValue); err != nil {
			utils.Warn("插入Host失败: %s -> %s, 错误: %v", h.Hostname, h.IP, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		utils.Debug("没有有效的Host配置被同步")
		return nil
	}

	// 替换全局DefaultHosts
	resolver.DefaultHosts = resolver.NewHosts(newHosts)
	utils.Info("Host同步完成: 共 %d 条配置，成功 %d 条", len(hosts), successCount)

	return nil
}

// ClearHosts 清空mihomo的自定义Hosts配置
func ClearHosts() {
	resolver.DefaultHosts = resolver.NewHosts(trie.New[resolver.HostValue]())
	utils.Debug("已清空mihomo Hosts配置")
}
