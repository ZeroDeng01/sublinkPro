package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"sublink/database"
	"sublink/node/protocol"
	"sublink/utils"

	"gorm.io/gorm"
)

// md5Hash 生成MD5哈希值（用于迁移老链接）
func md5Hash(src string) string {
	m := md5.New()
	m.Write([]byte(src))
	return hex.EncodeToString(m.Sum(nil))
}

// RunMigrations 执行所有数据库迁移
// 此函数必须在 database.InitSqlite() 之后调用
func RunMigrations() {
	db := database.DB
	if db == nil {
		utils.Error("数据库未初始化，无法执行迁移")
		return
	}

	// 检查是否已经初始化
	if database.IsInitialized {
		utils.Info("数据库已经初始化，无需重复初始化")
		return
	}

	// 基础数据库初始化
	if err := db.AutoMigrate(&User{}); err != nil {
		utils.Error("基础数据表User迁移失败: %v", err)
	} else {
		utils.Info("数据表User创建成功")
	}
	if err := db.AutoMigrate(&Subcription{}); err != nil {
		utils.Error("基础数据表Subcription迁移失败: %v", err)
	} else {
		utils.Info("数据表Subcription创建成功")
	}
	if err := db.AutoMigrate(&Node{}); err != nil {
		utils.Error("基础数据表Node迁移失败: %v", err)
	} else {
		utils.Info("数据表Node创建成功")
	}
	if err := db.AutoMigrate(&SubLogs{}); err != nil {
		utils.Error("基础数据表SubLogs迁移失败: %v", err)
	} else {
		utils.Info("数据表SubLogs创建成功")
	}
	if err := db.AutoMigrate(&AccessKey{}); err != nil {
		utils.Error("基础数据表AccessKey迁移失败: %v", err)
	} else {
		utils.Info("数据表AccessKey创建成功")
	}
	//if err := db.AutoMigrate(&SubScheduler{}); err != nil {
	//	utils.Error("基础数据表SubScheduler迁移失败: %v", err)
	//} else {
	//	utils.Info("数据表SubScheduler创建成功")
	//}
	if err := db.AutoMigrate(&SystemSetting{}); err != nil {
		utils.Error("基础数据表SystemSetting迁移失败: %v", err)
	} else {
		utils.Info("数据表SystemSetting创建成功")
	}
	if err := db.AutoMigrate(&Script{}); err != nil {
		utils.Error("基础数据表Script迁移失败: %v", err)
	} else {
		utils.Info("数据表Script创建成功")
	}
	if err := db.AutoMigrate(&SubcriptionGroup{}); err != nil {
		utils.Error("基础数据表SubcriptionGroup迁移失败: %v", err)
	} else {
		utils.Info("数据表SubcriptionGroup创建成功")
	}
	/*
		if err := db.AutoMigrate(&SubcriptionNode{}); err != nil {
			utils.Error("基础数据表SubcriptionNode迁移失败: %v", err)
		} else {
			utils.Info("数据表SubcriptionNode创建成功")
		}
	*/
	if err := db.AutoMigrate(&SubcriptionScript{}); err != nil {
		utils.Error("基础数据表SubcriptionScript迁移失败: %v", err)
	} else {
		utils.Info("数据表SubcriptionScript创建成功")
	}
	if err := db.AutoMigrate(&Template{}); err != nil {
		utils.Error("基础数据表Template迁移失败: %v", err)
	} else {
		utils.Info("数据表Template创建成功")
	}
	if err := db.AutoMigrate(&Tag{}); err != nil {
		utils.Error("基础数据表Tag迁移失败: %v", err)
	} else {
		utils.Info("数据表Tag创建成功")
	}
	if err := db.AutoMigrate(&TagRule{}); err != nil {
		utils.Error("基础数据表TagRule迁移失败: %v", err)
	} else {
		utils.Info("数据表TagRule创建成功")
	}
	if err := db.AutoMigrate(&Task{}); err != nil {
		utils.Error("基础数据表Task迁移失败: %v", err)
	} else {
		utils.Info("数据表Task创建成功")
	}
	if err := db.AutoMigrate(&IPInfo{}); err != nil {
		utils.Error("基础数据表IPInfo迁移失败: %v", err)
	} else {
		utils.Info("数据表IPInfo创建成功")
	}
	if err := db.AutoMigrate(&RememberToken{}); err != nil {
		utils.Error("基础数据表RememberToken迁移失败: %v", err)
	} else {
		utils.Info("数据表RememberToken创建成功")
	}
	if err := db.AutoMigrate(&Host{}); err != nil {
		utils.Error("基础数据表Host迁移失败: %v", err)
	} else {
		utils.Info("数据表Host创建成功")
	}
	if err := db.AutoMigrate(&SubscriptionShare{}); err != nil {
		utils.Error("基础数据表SubscriptionShare迁移失败: %v", err)
	} else {
		utils.Info("数据表SubscriptionShare创建成功")
	}
	if err := db.AutoMigrate(&SubscriptionChainRule{}); err != nil {
		utils.Error("基础数据表SubscriptionChainRule迁移失败: %v", err)
	} else {
		utils.Info("数据表SubscriptionChainRule创建成功")
	}
	if err := db.AutoMigrate(&Airport{}); err != nil {
		utils.Error("基础数据表Airport迁移失败: %v", err)
	} else {
		utils.Info("数据表Airport创建成功")
	}
	if err := db.AutoMigrate(&NodeCheckProfile{}); err != nil {
		utils.Error("基础数据表NodeCheckProfile迁移失败: %v", err)
	} else {
		utils.Info("数据表NodeCheckProfile创建成功")
	}

	// 检查并删除 idx_name_id 索引
	// 0000_drop_idx_name_id
	if err := database.RunCustomMigration("0000_drop_idx_name_id", func() error {
		if db.Migrator().HasIndex(&Node{}, "idx_name_id") {
			if err := db.Migrator().DropIndex(&Node{}, "idx_name_id"); err != nil {
				utils.Error("删除索引 idx_name_id 失败: %v", err)
				return err
			} else {
				utils.Info("成功删除索引 idx_name_id")
			}
		}
		return nil
	}); err != nil {
		utils.Error("执行迁移 0000_drop_idx_name_id 失败: %v", err)
	}

	// 0008_node_created_at_fill - 补全空的 CreatedAt 字段
	if err := database.RunCustomMigration("0008_node_created_at_fill", func() error {
		// 查找所有 CreatedAt 为零值的节点并设置为当前时间
		result := db.Exec("UPDATE nodes SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL OR created_at = '' OR created_at = '0001-01-01 00:00:00+00:00'")
		if result.Error != nil {
			return result.Error
		}
		utils.Info("已补全 %d 个节点的创建时间", result.RowsAffected)
		return nil
	}); err != nil {
		utils.Error("执行迁移 0008_node_created_at_fill 失败: %v", err)
	}

	// 0005_hash_passwords
	if err := database.RunCustomMigration("0005_hash_passwords", func() error {
		var users []User
		if err := db.Find(&users).Error; err != nil {
			return err
		}
		for _, user := range users {
			hashedPassword, err := HashPassword(user.Password)
			if err != nil {
				utils.Error("Failed to hash password for user %s: %v", user.Username, err)
				continue
			}
			user.Password = hashedPassword
			if err := db.Save(&user).Error; err != nil {
				utils.Error("Failed to save hashed password for user %s: %v", user.Username, err)
			} else {
				utils.Info("Upgraded password for user %s", user.Username)
			}
		}
		return nil
	}); err != nil {
		utils.Error("执行迁移 0005_hash_passwords 失败: %v", err)
	}

	// 添加脚本demo
	// 0007_add_script_demo
	if err := database.RunCustomMigration("0007_add_script_demo", func() error {
		script := &Script{}
		script.Name = "[系统DEMO]按测速结果筛选节点"
		script.Content = "" +
			"//修改节点列表\n/**\n * @param {Node[]} nodes\n * @param {string} clientType\n */\nfunction filterNode(nodes, clientType) {\n    let maxDelayTime = 250;//最大延迟 单位ms \n    let minSpeed = 1.5;//最小速度 单位MB/s\n    // nodes: 节点列表\n    // 数据结构如下\n    // [\n    //     {\n    //         \"ID\": 1,\n    //         \"Link\": \"vmess://4564564646\",\n    //         \"Name\": \"xx订阅_US-CDN-SSL\",\n    //         \"LinkName\": \"US-CDN-SSL\",\n    //         \"LinkAddress\": \"xxxxxxxxx.net:443\",\n    //         \"LinkHost\": \"xxxxxxxxx.net\",\n    //         \"LinkPort\": \"443\",\n    //         \"DialerProxyName\": \"\",\n    //         \"CreateDate\": \"\",\n    //         \"Source\": \"manual\",\n    //         \"SourceID\": 0,\n    //         \"Group\": \"自用\",\n    //         \"DelayTime\": 110,\n    //         \"Speed\": 10,\n    //         \"LastCheck\": \"2025-11-26 23:49:58\"\n    //     }\n    // ]\n    // clientType: 客户端类型\n    // 返回值: 修改后节点列表\n    let newNodes = [];\n    nodes.forEach(node => {\n        if(!node.Link.includes(\"://_\")){\n            //如果分组是机场或者自用的自建节点则忽略测速直接加入列表\n            if(node.Group.includes(\"机场\")||node.Group.includes(\"自建\")){\n                newNodes.push(rename(node));\n            }else{\n                //速度高或者延迟低都保留\n                if(node.DelayTime>0&&(node.DelayTime<maxDelayTime||node.Speed>=minSpeed)){\n                    newNodes.push(rename(node));\n                    console.log(\"✅节点：\"+node.Name +\"符合测速要求\");\n                }else{\n                    console.log(\"❌节点：\"+node.Name +\"不符合测速要求\");\n                }\n            }\n        }\n    });\n    return newNodes;\n}\n//修改订阅文件\n/**\n * @param {string} input\n * @param {string} clientType\n */\nfunction subMod( input, clientType) {\n    // input: 原始输入内容,不同客户端订阅文件也不一样\n    // clientType: 客户端类型\n    // 返回值: 修改后的内容字符串\n    return input; // 注意：此处示例仅为示意，实际应返回处理后的字符串\n}\n\n// 节点改名\nfunction rename(node){\n    if(node.Link.indexOf('#')!=-1&&node.Source!=='manual'){\n        var linkArr = node.Link.split('#')\n        node.Link = linkArr[0]+'#'+node.Source+\"_\"+linkArr[1]\n        return node\n    }\n\n    return node\n}"
		script.Version = "1.0.0"
		if script.CheckNameVersion() {
			return nil
		}
		err := db.First(&Script{}).Error
		if err == gorm.ErrRecordNotFound {
			err := script.Add()
			if err != nil {
				utils.Error("增加脚本demo失败: %v", err)
			}
		}
		return nil
	}); err != nil {
		utils.Error("执行迁移 0007_add_script_demo 失败: %v", err)
	}

	// 0009_migrate_template_files - 迁移现有模板文件到数据库
	if err := database.RunCustomMigration("0009_migrate_template_files", func() error {
		return MigrateTemplatesFromFiles("./template")
	}); err != nil {
		utils.Error("执行迁移 0009_migrate_template_files 失败: %v", err)
	}

	// 0010_add_default_base_templates - 添加默认基础模板到系统设置
	if err := database.RunCustomMigration("0010_add_default_base_templates", func() error {
		// 默认 Clash 模板
		clashTemplate := `port: 7890
socks-port: 7891
allow-lan: true
mode: Rule
log-level: info
external-controller: :9090
dns:
  enabled: true
  nameserver:
    - 119.29.29.29
    - 223.5.5.5
  fallback:
    - 8.8.8.8
    - 8.8.4.4
    - tls://1.0.0.1:853
    - tls://dns.google:853
proxies: ~

`
		// 默认 Surge 模板
		surgeTemplate := `[General]
loglevel = notify
bypass-system = true
skip-proxy = 127.0.0.1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,100.64.0.0/10,localhost,*.local,e.crashlytics.com,captive.apple.com,::ffff:0:0:0:0/1,::ffff:128:0:0:0/1
bypass-tun = 192.168.0.0/16,10.0.0.0/8,172.16.0.0/12
dns-server = 119.29.29.29,223.5.5.5,218.30.19.40,61.134.1.4
external-controller-access = password@0.0.0.0:6170
http-api = password@0.0.0.0:6171
test-timeout = 5
http-api-web-dashboard = true
exclude-simple-hostnames = true
allow-wifi-access = true
http-listen = 0.0.0.0:6152
socks5-listen = 0.0.0.0:6153
wifi-access-http-port = 6152
wifi-access-socks5-port = 6153

[Proxy]
DIRECT = direct

`
		// 插入 Clash 模板
		if err := SetSetting("base_template_clash", clashTemplate); err != nil {
			utils.Error("插入 Clash 基础模板失败: %v", err)
			return err
		}
		// 插入 Surge 模板
		if err := SetSetting("base_template_surge", surgeTemplate); err != nil {
			utils.Error("插入 Surge 基础模板失败: %v", err)
			return err
		}
		utils.Info("已添加默认 Clash 和 Surge 基础模板")
		return nil
	}); err != nil {
		utils.Error("执行迁移 0010_add_default_base_templates 失败: %v", err)
	}

	// 0011_migrate_speed_test_concurrency - 迁移旧的并发数配置到新的分离配置
	if err := database.RunCustomMigration("0011_migrate_speed_test_concurrency", func() error {
		// 读取旧的 speed_test_concurrency 配置
		oldConcurrency, _ := GetSetting("speed_test_concurrency")
		if oldConcurrency != "" {
			// 将旧配置迁移到 latency_concurrency
			if err := SetSetting("speed_test_latency_concurrency", oldConcurrency); err != nil {
				utils.Error("迁移 latency_concurrency 失败: %v", err)
				return err
			}
			utils.Info("已将 speed_test_concurrency=%s 迁移到 speed_test_latency_concurrency", oldConcurrency)
		}

		// 设置默认的 speed_concurrency 为 1（如果不存在）
		existingSpeedConcurrency, _ := GetSetting("speed_test_speed_concurrency")
		if existingSpeedConcurrency == "" {
			if err := SetSetting("speed_test_speed_concurrency", "1"); err != nil {
				utils.Error("设置默认 speed_concurrency 失败: %v", err)
				return err
			}
			utils.Info("已设置默认 speed_test_speed_concurrency=1")
		}

		// 设置默认的 latency_samples 为 3（如果不存在）
		existingLatencySamples, _ := GetSetting("speed_test_latency_samples")
		if existingLatencySamples == "" {
			if err := SetSetting("speed_test_latency_samples", "3"); err != nil {
				utils.Error("设置默认 latency_samples 失败: %v", err)
				return err
			}
			utils.Info("已设置默认 speed_test_latency_samples=3")
		}

		return nil
	}); err != nil {
		utils.Error("执行迁移 0011_migrate_speed_test_concurrency 失败: %v", err)
	}

	// 0012_migrate_last_check_to_separate_fields - 将 LastCheck 字段迁移到 LatencyCheckAt 和 SpeedCheckAt
	if err := database.RunCustomMigration("0012_migrate_last_check_to_separate_fields", func() error {
		// 检查 last_check 列是否存在
		if db.Migrator().HasColumn(&Node{}, "last_check") {
			// 将 last_check 数据复制到 latency_check_at 和 speed_check_at
			result := db.Exec("UPDATE nodes SET latency_check_at = last_check, speed_check_at = last_check WHERE last_check IS NOT NULL AND last_check != ''")
			if result.Error != nil {
				utils.Error("迁移 last_check 数据失败: %v", result.Error)
				return result.Error
			}
			utils.Info("已将 %d 条 last_check 数据迁移到新字段", result.RowsAffected)

			// 删除 last_check 列
			if err := db.Exec("ALTER TABLE nodes DROP COLUMN last_check").Error; err != nil {
				utils.Error("删除 last_check 列失败: %v", err)
				// 不返回错误，因为某些数据库可能不支持 DROP COLUMN
			} else {
				utils.Info("成功删除 last_check 列")
			}
		}
		return nil
	}); err != nil {
		utils.Error("执行迁移 0012_migrate_last_check_to_separate_fields 失败: %v", err)
	}

	// 0013_migrate_node_status_fields - 根据已有数据设置 SpeedStatus 和 DelayStatus 字段
	if err := database.RunCustomMigration("0013_migrate_node_status_fields", func() error {
		// DelayTime > 0 且有记录 => DelayStatus = 'success'
		if result := db.Exec("UPDATE nodes SET delay_status = 'success' WHERE delay_time > 0 AND (delay_status IS NULL OR delay_status = '' OR delay_status = 'untested')"); result.Error != nil {
			utils.Error("迁移 DelayStatus (success) 失败: %v", result.Error)
		} else {
			utils.Info("已设置 %d 个节点 DelayStatus 为 success", result.RowsAffected)
		}

		// DelayTime = -1 => DelayStatus = 'timeout'
		if result := db.Exec("UPDATE nodes SET delay_status = 'timeout' WHERE delay_time = -1 AND (delay_status IS NULL OR delay_status = '' OR delay_status = 'untested')"); result.Error != nil {
			utils.Error("迁移 DelayStatus (timeout) 失败: %v", result.Error)
		} else {
			utils.Info("已设置 %d 个节点 DelayStatus 为 timeout", result.RowsAffected)
		}

		// Speed > 0 => SpeedStatus = 'success'
		if result := db.Exec("UPDATE nodes SET speed_status = 'success' WHERE speed > 0 AND (speed_status IS NULL OR speed_status = '' OR speed_status = 'untested')"); result.Error != nil {
			utils.Error("迁移 SpeedStatus (success) 失败: %v", result.Error)
		} else {
			utils.Info("已设置 %d 个节点 SpeedStatus 为 success", result.RowsAffected)
		}

		// Speed = -1 => SpeedStatus = 'error'
		if result := db.Exec("UPDATE nodes SET speed_status = 'error' WHERE speed = -1 AND (speed_status IS NULL OR speed_status = '' OR speed_status = 'untested')"); result.Error != nil {
			utils.Error("迁移 SpeedStatus (error) 失败: %v", result.Error)
		} else {
			utils.Info("已设置 %d 个节点 SpeedStatus 为 error", result.RowsAffected)
		}

		// 所有其他情况 => 'untested'
		if result := db.Exec("UPDATE nodes SET speed_status = 'untested' WHERE speed_status IS NULL OR speed_status = ''"); result.Error != nil {
			utils.Error("迁移 SpeedStatus (untested) 失败: %v", result.Error)
		}
		if result := db.Exec("UPDATE nodes SET delay_status = 'untested' WHERE delay_status IS NULL OR delay_status = ''"); result.Error != nil {
			utils.Error("迁移 DelayStatus (untested) 失败: %v", result.Error)
		}

		utils.Info("节点状态字段迁移完成")
		return nil
	}); err != nil {
		utils.Error("执行迁移 0013_migrate_node_status_fields 失败: %v", err)
	}

	// 0014_migrate_subcription_node_to_id_v2 - 将 SubcriptionNode 表从 NodeName 关联改为 NodeID 关联 (v2 强制重试)
	if err := database.RunCustomMigration("0014_migrate_subcription_node_to_id_v2", func() error {
		// 0. 如果表不存在，直接创建新表
		if !db.Migrator().HasTable(&SubcriptionNode{}) {
			if err := db.AutoMigrate(&SubcriptionNode{}); err != nil {
				return fmt.Errorf("创建新表失败: %w", err)
			}
			utils.Info("创建了新的 SubcriptionNode 表")
			return nil
		}

		// 检查 node_name 列是否存在（判断是否需要迁移）
		// 注意：不能使用 &SubcriptionNode{} 检查，因为结构体已修改
		if !db.Migrator().HasColumn("subcription_nodes", "node_name") {
			utils.Info("SubcriptionNode 表已是 NodeID 关联，无需迁移")
			// 确保 node_id 列存在（针对某些异常情况）
			if !db.Migrator().HasColumn("subcription_nodes", "node_id") {
				return db.AutoMigrate(&SubcriptionNode{})
			}
			return nil
		}

		utils.Info("开始迁移 SubcriptionNode 表从 NodeName 到 NodeID...")

		// 1. 备份原表 (如果存在先删除)
		_ = db.Exec("DROP TABLE IF EXISTS subcription_nodes_backup")
		if err := db.Exec("CREATE TABLE subcription_nodes_backup AS SELECT * FROM subcription_nodes").Error; err != nil {
			utils.Warn("备份表创建失败: %v", err)
			return fmt.Errorf("备份表失败: %w", err)
		} else {
			utils.Info("已创建备份表 subcription_nodes_backup")
		}

		// 2. 添加 node_id 列（如果不存在）
		if !db.Migrator().HasColumn(&SubcriptionNode{}, "node_id") {
			if err := db.Exec("ALTER TABLE subcription_nodes ADD COLUMN node_id INTEGER").Error; err != nil {
				return fmt.Errorf("添加 node_id 列失败: %w", err)
			}
		}

		// 3. 通过 JOIN 更新 node_id
		result := db.Exec(`
			UPDATE subcription_nodes 
			SET node_id = (
				SELECT nodes.id FROM nodes 
				WHERE nodes.name = subcription_nodes.node_name
				LIMIT 1
			)
			WHERE node_id IS NULL OR node_id = 0
		`)
		if result.Error != nil {
			return fmt.Errorf("更新 node_id 失败: %w", result.Error)
		}
		utils.Info("已更新 %d 条记录的 node_id", result.RowsAffected)

		// 4. 清理无效关联（node_name 对应的节点已不存在）
		cleanResult := db.Exec("DELETE FROM subcription_nodes WHERE node_id IS NULL OR node_id = 0")
		if cleanResult.Error != nil {
			utils.Warn("清理无效关联失败: %v", cleanResult.Error)
		} else if cleanResult.RowsAffected > 0 {
			utils.Info("已清理 %d 条无效关联（节点已删除）", cleanResult.RowsAffected)
		}

		// 5. 重建表（SQLite 不支持 DROP COLUMN）
		if err := db.Exec(`
			CREATE TABLE subcription_nodes_new (
				subcription_id INTEGER NOT NULL,
				node_id INTEGER NOT NULL,
				sort INTEGER DEFAULT 0,
				PRIMARY KEY (subcription_id, node_id)
			)
		`).Error; err != nil {
			return fmt.Errorf("创建新表失败: %w", err)
		}

		if err := db.Exec(`
			INSERT INTO subcription_nodes_new (subcription_id, node_id, sort)
			SELECT subcription_id, node_id, sort FROM subcription_nodes
			WHERE node_id IS NOT NULL AND node_id > 0
		`).Error; err != nil {
			return fmt.Errorf("迁移数据失败: %w", err)
		}

		if err := db.Exec("DROP TABLE subcription_nodes").Error; err != nil {
			return fmt.Errorf("删除旧表失败: %w", err)
		}

		if err := db.Exec("ALTER TABLE subcription_nodes_new RENAME TO subcription_nodes").Error; err != nil {
			return fmt.Errorf("重命名表失败: %w", err)
		}

		utils.Info("SubcriptionNode 表迁移完成")
		return nil
	}); err != nil {
		utils.Error("执行迁移 0014_migrate_subcription_node_to_id_v2 失败: %v", err)
	}

	// 0015_migrate_subscription_shares - 将老订阅的MD5分享链接迁移到新的分享表
	if err := database.RunCustomMigration("0015_migrate_subscription_shares", func() error {
		// 获取所有订阅
		var subs []Subcription
		if err := db.Find(&subs).Error; err != nil {
			return fmt.Errorf("获取订阅列表失败: %w", err)
		}

		migratedCount := 0
		logsUpdatedCount := 0
		for _, sub := range subs {
			// 检查该订阅是否已有分享记录
			var existingCount int64
			db.Model(&SubscriptionShare{}).Where("subscription_id = ? AND is_legacy = ?", sub.ID, true).Count(&existingCount)
			if existingCount > 0 {
				continue // 已迁移过，跳过
			}

			// 生成老的 MD5 token
			token := md5Hash(sub.Name)

			// 创建分享记录
			share := SubscriptionShare{
				SubscriptionID: sub.ID,
				Token:          token,
				Name:           "默认分享链接",
				ExpireType:     ExpireTypeNever, // 永不过期
				IsLegacy:       true,
				Enabled:        true,
			}

			if err := db.Create(&share).Error; err != nil {
				utils.Warn("迁移订阅 %s 的分享链接失败: %v", sub.Name, err)
				continue
			}
			migratedCount++

			// 将该订阅下 ShareID=0 的老访问日志关联到新创建的默认分享链接
			result := db.Model(&SubLogs{}).
				Where("subcription_id = ? AND (share_id = 0 OR share_id IS NULL)", sub.ID).
				Update("share_id", share.ID)
			if result.Error != nil {
				utils.Warn("更新订阅 %s 的访问日志失败: %v", sub.Name, result.Error)
			} else if result.RowsAffected > 0 {
				logsUpdatedCount += int(result.RowsAffected)
			}
		}

		utils.Info("已为 %d 个订阅创建默认分享链接，更新了 %d 条访问日志", migratedCount, logsUpdatedCount)
		return nil
	}); err != nil {
		utils.Error("执行迁移 0015_migrate_subscription_shares 失败: %v", err)
	}

	// 0016_add_node_protocol_field - 为现有节点填充协议类型字段
	if err := database.RunCustomMigration("0016_add_node_protocol_field", func() error {
		// 获取所有节点的 ID 和 Link
		var nodes []struct {
			ID   int
			Link string
		}
		if err := db.Model(&Node{}).Select("id", "link").Find(&nodes).Error; err != nil {
			return fmt.Errorf("获取节点列表失败: %w", err)
		}

		if len(nodes) == 0 {
			utils.Info("没有需要迁移的节点")
			return nil
		}

		// 按协议类型分组，减少 SQL 执行次数
		protoGroups := make(map[string][]int)
		for _, node := range nodes {
			protoType := protocol.GetProtocolFromLink(node.Link)
			protoGroups[protoType] = append(protoGroups[protoType], node.ID)
		}

		// 批量更新每组
		for protoType, ids := range protoGroups {
			if err := db.Model(&Node{}).Where("id IN ?", ids).Update("protocol", protoType).Error; err != nil {
				utils.Warn("批量更新协议类型 %s 失败: %v", protoType, err)
			}
		}

		utils.Info("已为 %d 个节点填充协议类型字段，共 %d 种协议", len(nodes), len(protoGroups))
		return nil
	}); err != nil {
		utils.Error("执行迁移 0016_add_node_protocol_field 失败: %v", err)
	}

	// 0017_migrate_subscheduler_to_airport - 将SubScheduler数据迁移到Airport表
	if err := database.RunCustomMigration("0017_migrate_subscheduler_to_airport", func() error {
		// 检查旧表是否存在
		if !db.Migrator().HasTable("sub_schedulers") {
			utils.Info("SubScheduler表不存在，无需迁移")
			return nil
		}

		// 检查新表是否为空（仅空表时才迁移，避免重复迁移）
		var airportCount int64
		db.Model(&Airport{}).Count(&airportCount)
		if airportCount > 0 {
			utils.Info("Airport表已有数据，跳过迁移")
			return nil
		}

		// 获取所有SubScheduler数据
		var schedulers []SubScheduler
		if err := db.Find(&schedulers).Error; err != nil {
			return fmt.Errorf("获取SubScheduler数据失败: %w", err)
		}

		if len(schedulers) == 0 {
			utils.Info("SubScheduler表为空，无需迁移")
			return nil
		}

		// 迁移数据到Airport表
		for _, s := range schedulers {
			airport := Airport{
				ID:                s.ID,
				Name:              s.Name,
				URL:               s.URL,
				CronExpr:          s.CronExpr,
				Enabled:           s.Enabled,
				SuccessCount:      s.SuccessCount,
				LastRunTime:       s.LastRunTime,
				NextRunTime:       s.NextRunTime,
				CreatedAt:         s.CreatedAt,
				UpdatedAt:         s.UpdatedAt,
				Group:             s.Group,
				DownloadWithProxy: s.DownloadWithProxy,
				ProxyLink:         s.ProxyLink,
				UserAgent:         s.UserAgent,
			}
			if err := db.Create(&airport).Error; err != nil {
				utils.Warn("迁移机场 %s 失败: %v", s.Name, err)
				continue
			}
		}

		utils.Info("已将 %d 个SubScheduler记录迁移到Airport表", len(schedulers))
		return nil
	}); err != nil {
		utils.Error("执行迁移 0017_migrate_subscheduler_to_airport 失败: %v", err)
	}

	// 0018_migrate_speed_test_to_node_check_profile - 迁移测速配置到节点检测策略表
	if err := database.RunCustomMigration("0018_migrate_speed_test_to_node_check_profile", func() error {
		// 检查是否已有策略记录
		var count int64
		db.Model(&NodeCheckProfile{}).Count(&count)
		if count > 0 {
			utils.Info("节点检测策略表已有数据，跳过迁移")
			return nil
		}

		// 从 system_settings 读取现有测速配置
		cron, _ := GetSetting("speed_test_cron")
		enabledStr, _ := GetSetting("speed_test_enabled")
		enabled := enabledStr == "true"
		mode, _ := GetSetting("speed_test_mode")
		if mode == "" {
			mode = "tcp"
		}
		testURL, _ := GetSetting("speed_test_url")
		latencyURL, _ := GetSetting("speed_test_latency_url")
		timeoutStr, _ := GetSetting("speed_test_timeout")
		timeout := 5
		if timeoutStr != "" {
			if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
				timeout = t
			}
		}
		groups, _ := GetSetting("speed_test_groups")
		tags, _ := GetSetting("speed_test_tags")
		latencyConcurrencyStr, _ := GetSetting("speed_test_latency_concurrency")
		latencyConcurrency := 0
		if latencyConcurrencyStr != "" {
			latencyConcurrency, _ = strconv.Atoi(latencyConcurrencyStr)
		}
		speedConcurrencyStr, _ := GetSetting("speed_test_speed_concurrency")
		speedConcurrency := 1
		if speedConcurrencyStr != "" {
			if c, err := strconv.Atoi(speedConcurrencyStr); err == nil && c > 0 {
				speedConcurrency = c
			}
		}
		detectCountryStr, _ := GetSetting("speed_test_detect_country")
		detectCountry := detectCountryStr == "true"
		landingIPURL, _ := GetSetting("speed_test_landing_ip_url")
		includeHandshakeStr, _ := GetSetting("speed_test_include_handshake")
		includeHandshake := includeHandshakeStr != "false"
		speedRecordMode, _ := GetSetting("speed_test_speed_record_mode")
		if speedRecordMode == "" {
			speedRecordMode = "average"
		}
		peakSampleIntervalStr, _ := GetSetting("speed_test_peak_sample_interval")
		peakSampleInterval := 100
		if peakSampleIntervalStr != "" {
			if v, err := strconv.Atoi(peakSampleIntervalStr); err == nil && v >= 50 && v <= 200 {
				peakSampleInterval = v
			}
		}
		trafficByGroupStr, _ := GetSetting("speed_test_traffic_by_group")
		trafficByGroup := trafficByGroupStr != "false"
		trafficBySourceStr, _ := GetSetting("speed_test_traffic_by_source")
		trafficBySource := trafficBySourceStr != "false"
		trafficByNodeStr, _ := GetSetting("speed_test_traffic_by_node")
		trafficByNode := trafficByNodeStr == "true"

		// 创建默认策略
		defaultProfile := NodeCheckProfile{
			Name:               "默认策略",
			Enabled:            enabled,
			CronExpr:           cron,
			Mode:               mode,
			TestURL:            testURL,
			LatencyURL:         latencyURL,
			Timeout:            timeout,
			Groups:             groups,
			Tags:               tags,
			LatencyConcurrency: latencyConcurrency,
			SpeedConcurrency:   speedConcurrency,
			DetectCountry:      detectCountry,
			LandingIPURL:       landingIPURL,
			IncludeHandshake:   includeHandshake,
			SpeedRecordMode:    speedRecordMode,
			PeakSampleInterval: peakSampleInterval,
			TrafficByGroup:     trafficByGroup,
			TrafficBySource:    trafficBySource,
			TrafficByNode:      trafficByNode,
		}

		if err := db.Create(&defaultProfile).Error; err != nil {
			return fmt.Errorf("创建默认节点检测策略失败: %w", err)
		}

		utils.Info("已将现有测速配置迁移到默认节点检测策略")
		return nil
	}); err != nil {
		utils.Error("执行迁移 0018_migrate_speed_test_to_node_check_profile 失败: %v", err)
	}

	// 0019_fill_empty_node_protocol - 为 protocol 为空的节点补充协议类型
	if err := database.RunCustomMigration("0019_fill_empty_node_protocol", func() error {
		// 查找所有 protocol 为空的节点
		var nodes []struct {
			ID   int
			Link string
		}
		if err := db.Model(&Node{}).
			Select("id", "link").
			Where("protocol IS NULL OR protocol = ''").
			Find(&nodes).Error; err != nil {
			return fmt.Errorf("查询 protocol 为空的节点失败: %w", err)
		}

		if len(nodes) == 0 {
			utils.Info("没有 protocol 为空的节点需要处理")
			return nil
		}

		// 按协议类型分组，减少 SQL 执行次数
		protoGroups := make(map[string][]int)
		for _, node := range nodes {
			protoType := protocol.GetProtocolFromLink(node.Link)
			protoGroups[protoType] = append(protoGroups[protoType], node.ID)
		}

		// 批量更新每组
		updateCount := 0
		for protoType, ids := range protoGroups {
			if err := db.Model(&Node{}).Where("id IN ?", ids).Update("protocol", protoType).Error; err != nil {
				utils.Warn("批量更新协议类型 %s 失败: %v", protoType, err)
			} else {
				updateCount += len(ids)
			}
		}

		utils.Info("已为 %d 个 protocol 为空的节点补充协议类型，共 %d 种协议", updateCount, len(protoGroups))
		return nil
	}); err != nil {
		utils.Error("执行迁移 0019_fill_empty_node_protocol 失败: %v", err)
	}

	// 初始化用户数据
	err := db.First(&User{}).Error
	if err == gorm.ErrRecordNotFound {
		adminPassword := "123456"
		if envPass := os.Getenv("SUBLINK_ADMIN_PASSWORD"); envPass != "" {
			adminPassword = envPass
		}
		admin := &User{
			Username: "admin",
			Password: adminPassword,
			Role:     "admin",
			Nickname: "管理员",
		}
		err = admin.Create()
		if err != nil {
			utils.Error("初始化添加用户数据失败")
		}
	} else {
		// Check if we need to update admin password from env
		if envPass := os.Getenv("SUBLINK_ADMIN_PASSWORD_REST"); envPass != "" {
			var admin User
			if err := db.First(&admin).Error; err == nil {
				// Update admin password
				updateUser := &User{Password: envPass}
				if err := admin.Set(updateUser); err != nil {
					utils.Error("Failed to update admin password from env: %v", err)
				} else {
					utils.Info("Admin password updated from environment variable")
				}
			}
		}
	}

	// 设置初始化标志为 true
	database.IsInitialized = true
	utils.Info("数据库初始化成功")
}

// Rollback0014_migrate_subcription_node_to_id_v2 回滚迁移 0014
// 此函数需要手动调用，用于出现问题时回滚
func Rollback0014_migrate_subcription_node_to_id_v2() error {
	db := database.DB
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	utils.Info("开始回滚 SubcriptionNode 表迁移...")

	// 检查备份表是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='subcription_nodes_backup'").Scan(&count)
	if count == 0 {
		return fmt.Errorf("备份表 subcription_nodes_backup 不存在，无法回滚")
	}

	// 1. 删除当前表
	if err := db.Exec("DROP TABLE IF EXISTS subcription_nodes").Error; err != nil {
		return fmt.Errorf("删除当前表失败: %w", err)
	}

	// 2. 从备份恢复
	if err := db.Exec("CREATE TABLE subcription_nodes AS SELECT * FROM subcription_nodes_backup").Error; err != nil {
		return fmt.Errorf("从备份恢复失败: %w", err)
	}

	// 3. 删除迁移记录
	if err := db.Exec("DELETE FROM schema_migrations WHERE version = '0014_migrate_subcription_node_to_id_v2'").Error; err != nil {
		utils.Warn("删除迁移记录失败: %v", err)
	}

	utils.Info("回滚完成，已恢复到原表结构")
	return nil
}
