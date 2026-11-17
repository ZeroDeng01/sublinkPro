package models

import (
	"sublink/database"
	"sublink/utils"
	"time"
)

// InitDemoData 初始化演示数据
// 仅在演示模式下调用，用于填充演示用的测试数据
func InitDemoData() {
	if !IsDemoMode() {
		return
	}

	utils.Info("🎭 Initializing demo data...")

	// 创建演示用户
	initDemoUser()

	// 创建演示节点
	initDemoNodes()

	// 创建演示标签
	initDemoTags()

	// 创建演示订阅
	initDemoSubscriptions()

	// 刷新缓存
	refreshDemoCaches()

	utils.Info("🎭 Demo data initialized")
}

// initDemoUser 创建演示用户
func initDemoUser() {
	user := &User{
		Username: "admin",
		Password: "123456",
		Role:     "admin",
		Nickname: "Demo Admin",
	}
	if err := user.Create(); err != nil {
		utils.Error("Failed to create demo user: %v", err)
	} else {
		utils.Info("✓ Created demo user: admin")
	}
}

// initDemoNodes 创建演示节点
func initDemoNodes() {
	now := time.Now()
	demoNodes := []Node{
		{
			Name:        "🇭🇰 Demo-HK-01",
			LinkName:    "Demo-HK-01",
			Link:        "vmess://eyJhZGQiOiJkZW1vLWhvbmdrb25nMDEuZXhhbXBsZS5jb20iLCJhaWQiOiIwIiwiaG9zdCI6IiIsImlkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5YWJjIiwibmV0IjoidGNwIiwicGF0aCI6IiIsInBvcnQiOiI0NDMiLCJwcyI6IkRlbW8tSEstMDEiLCJzY3kiOiJhdXRvIiwic25pIjoiIiwidGxzIjoidGxzIiwidHlwZSI6Im5vbmUiLCJ2IjoiMiJ9",
			LinkAddress: "demo-hongkong01.example.com:443",
			LinkHost:    "demo-hongkong01.example.com",
			LinkPort:    "443",
			LinkCountry: "HK",
			Source:      "demo",
			Group:       "Demo Group",
			Speed:       25.6,
			DelayTime:   45,
			SpeedStatus: "success",
			DelayStatus: "success",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "🇯🇵 Demo-JP-01",
			LinkName:    "Demo-JP-01",
			Link:        "vmess://eyJhZGQiOiJkZW1vLWphcGFuMDEuZXhhbXBsZS5jb20iLCJhaWQiOiIwIiwiaG9zdCI6IiIsImlkIjoiMjM0NTY3ODktMjM0NS0yMzQ1LTIzNDUtMjM0NTY3ODlhYmNkIiwibmV0IjoidGNwIiwicGF0aCI6IiIsInBvcnQiOiI0NDMiLCJwcyI6IkRlbW8tSlAtMDEiLCJzY3kiOiJhdXRvIiwic25pIjoiIiwidGxzIjoidGxzIiwidHlwZSI6Im5vbmUiLCJ2IjoiMiJ9",
			LinkAddress: "demo-japan01.example.com:443",
			LinkHost:    "demo-japan01.example.com",
			LinkPort:    "443",
			LinkCountry: "JP",
			Source:      "demo",
			Group:       "Demo Group",
			Speed:       32.1,
			DelayTime:   68,
			SpeedStatus: "success",
			DelayStatus: "success",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "🇸🇬 Demo-SG-01",
			LinkName:    "Demo-SG-01",
			Link:        "trojan://demo-password-sg@demo-singapore01.example.com:443?security=tls#🇸🇬 Demo-SG-01",
			LinkAddress: "demo-singapore01.example.com:443",
			LinkHost:    "demo-singapore01.example.com",
			LinkPort:    "443",
			LinkCountry: "SG",
			Source:      "demo",
			Group:       "Demo Group",
			Speed:       18.5,
			DelayTime:   85,
			SpeedStatus: "success",
			DelayStatus: "success",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "🇺🇸 Demo-US-01",
			LinkName:    "Demo-US-01",
			Link:        "ss://YWVzLTI1Ni1nY206ZGVtby1wYXNzd29yZC11c0BkZW1vLXVzYTAxLmV4YW1wbGUuY29tOjQ0Mw==#🇺🇸 Demo-US-01",
			LinkAddress: "demo-usa01.example.com:443",
			LinkHost:    "demo-usa01.example.com",
			LinkPort:    "443",
			LinkCountry: "US",
			Source:      "demo",
			Group:       "Demo Group",
			Speed:       15.2,
			DelayTime:   180,
			SpeedStatus: "success",
			DelayStatus: "success",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "🇩🇪 Demo-DE-01",
			LinkName:    "Demo-DE-01",
			Link:        "vless://abcdef12-3456-7890-abcd-ef1234567890@demo-germany01.example.com:443?encryption=none&security=tls&type=ws&host=demo-germany01.example.com&path=/demo#🇩🇪 Demo-DE-01",
			LinkAddress: "demo-germany01.example.com:443",
			LinkHost:    "demo-germany01.example.com",
			LinkPort:    "443",
			LinkCountry: "DE",
			Source:      "demo",
			Group:       "Europe Nodes",
			Speed:       12.8,
			DelayTime:   210,
			SpeedStatus: "success",
			DelayStatus: "success",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "🇬🇧 Demo-UK-01",
			LinkName:    "Demo-UK-01",
			Link:        "hy2://demo-password-uk@demo-uk01.example.com:443?sni=demo-uk01.example.com#🇬🇧 Demo-UK-01",
			LinkAddress: "demo-uk01.example.com:443",
			LinkHost:    "demo-uk01.example.com",
			LinkPort:    "443",
			LinkCountry: "GB",
			Source:      "demo",
			Group:       "Europe Nodes",
			Speed:       0,
			DelayTime:   0,
			SpeedStatus: "untested",
			DelayStatus: "untested",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "🇰🇷 Demo-KR-01",
			LinkName:    "Demo-KR-01",
			Link:        "vmess://eyJhZGQiOiJkZW1vLWtvcmVhMDEuZXhhbXBsZS5jb20iLCJhaWQiOiIwIiwiaG9zdCI6IiIsImlkIjoiMzQ1Njc4OTAtMzQ1Ni0zNDU2LTM0NTYtMzQ1Njc4OTBhYmNkIiwibmV0Ijoid3MiLCJwYXRoIjoiL2RlbW8iLCJwb3J0IjoiNDQzIiwicHMiOiJEZW1vLUtSLTAxIiwic2N5IjoiYXV0byIsInNuaSI6IiIsInRscyI6InRscyIsInR5cGUiOiJub25lIiwidiI6IjIifQ==",
			LinkAddress: "demo-korea01.example.com:443",
			LinkHost:    "demo-korea01.example.com",
			LinkPort:    "443",
			LinkCountry: "KR",
			Source:      "demo",
			Group:       "Demo Group",
			Speed:       28.3,
			DelayTime:   52,
			SpeedStatus: "success",
			DelayStatus: "success",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "🇹🇼 Demo-TW-01",
			LinkName:    "Demo-TW-01",
			Link:        "trojan://demo-password-tw@demo-taiwan01.example.com:443?security=tls#🇹🇼 Demo-TW-01",
			LinkAddress: "demo-taiwan01.example.com:443",
			LinkHost:    "demo-taiwan01.example.com",
			LinkPort:    "443",
			LinkCountry: "TW",
			Source:      "demo",
			Group:       "Demo Group",
			Speed:       22.5,
			DelayTime:   58,
			SpeedStatus: "success",
			DelayStatus: "success",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for i := range demoNodes {
		if err := demoNodes[i].Add(); err != nil {
			utils.Error("Failed to create demo node: %v", err)
		} else {
			utils.Info("✓ Created demo node: %s", demoNodes[i].Name)
		}
	}
}

// initDemoTags 创建演示标签
func initDemoTags() {
	demoTags := []Tag{
		{Name: "Fast Nodes", Color: "#4CAF50"},
		{Name: "Stable Nodes", Color: "#2196F3"},
		{Name: "Streaming", Color: "#FF9800"},
		{Name: "Gaming", Color: "#9C27B0"},
	}

	for i := range demoTags {
		if err := demoTags[i].Add(); err != nil {
			utils.Error("Failed to create demo tag: %v", err)
		} else {
			utils.Info("✓ Created demo tag: %s", demoTags[i].Name)
		}
	}
}

// initDemoSubscriptions 创建演示订阅
func initDemoSubscriptions() {
	// 默认订阅配置 JSON（包含模板路径和选项）
	defaultConfig := `{"clash":"./template/clash.yaml","surge":"./template/surge.conf","udp":false,"cert":false}`

	// 创建一个基础订阅
	sub := &Subcription{
		Name:   "Demo Subscription - Global",
		Config: defaultConfig,
	}
	if err := sub.Add(); err != nil {
		utils.Error("Failed to create demo subscription: %v", err)
		return
	}

	// 添加分组关联
	if err := sub.AddGroups([]string{"Demo Group"}); err != nil {
		utils.Error("Failed to add subscription group: %v", err)
	}

	utils.Info("✓ Created demo subscription: %s", sub.Name)

	// 创建欧洲订阅
	subEurope := &Subcription{
		Name:   "Demo Subscription - Europe",
		Config: defaultConfig,
	}
	if err := subEurope.Add(); err != nil {
		utils.Error("Failed to create demo subscription: %v", err)
		return
	}

	if err := subEurope.AddGroups([]string{"Europe Nodes"}); err != nil {
		utils.Error("Failed to add subscription group: %v", err)
	}

	utils.Info("✓ Created demo subscription: %s", subEurope.Name)
}

// refreshDemoCaches 刷新演示数据相关的缓存
func refreshDemoCaches() {
	// 重新初始化缓存以确保演示数据被正确加载
	if err := InitNodeCache(); err != nil {
		utils.Error("Failed to refresh node cache: %v", err)
	}
	if err := InitTagCache(); err != nil {
		utils.Error("Failed to refresh tag cache: %v", err)
	}
	if err := InitSubcriptionCache(); err != nil {
		utils.Error("Failed to refresh subscription cache: %v", err)
	}
}

// ResetDemoData 重置演示数据（可选功能，用于定期重置）
func ResetDemoData() {
	if !IsDemoMode() {
		return
	}

	utils.Info("🔄 Resetting demo data...")

	// 清空所有表
	database.DB.Exec("DELETE FROM nodes")
	database.DB.Exec("DELETE FROM tags")
	database.DB.Exec("DELETE FROM subcriptions")
	database.DB.Exec("DELETE FROM subcription_groups")
	database.DB.Exec("DELETE FROM subcription_nodes")
	database.DB.Exec("DELETE FROM subcription_scripts")
	database.DB.Exec("DELETE FROM users")
	database.DB.Exec("DELETE FROM subcription_tags")

	// 重新初始化演示数据
	InitDemoData()
}
