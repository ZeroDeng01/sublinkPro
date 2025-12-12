package models

import (
	"log"
	"os"
	"sublink/database"

	"gorm.io/gorm"
)

// RunMigrations 执行所有数据库迁移
// 此函数必须在 database.InitSqlite() 之后调用
func RunMigrations() {
	db := database.DB
	if db == nil {
		log.Println("数据库未初始化，无法执行迁移")
		return
	}

	// 检查是否已经初始化
	if database.IsInitialized {
		log.Println("数据库已经初始化，无需重复初始化")
		return
	}

	// 基础数据库初始化
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Printf("基础数据表User迁移失败: %v", err)
	} else {
		log.Printf("数据表User创建成功")
	}
	if err := db.AutoMigrate(&Subcription{}); err != nil {
		log.Printf("基础数据表Subcription迁移失败: %v", err)
	} else {
		log.Printf("数据表Subcription创建成功")
	}
	if err := db.AutoMigrate(&Node{}); err != nil {
		log.Printf("基础数据表Node迁移失败: %v", err)
	} else {
		log.Printf("数据表Node创建成功")
	}
	if err := db.AutoMigrate(&SubLogs{}); err != nil {
		log.Printf("基础数据表SubLogs迁移失败: %v", err)
	} else {
		log.Printf("数据表SubLogs创建成功")
	}
	if err := db.AutoMigrate(&AccessKey{}); err != nil {
		log.Printf("基础数据表AccessKey迁移失败: %v", err)
	} else {
		log.Printf("数据表AccessKey创建成功")
	}
	if err := db.AutoMigrate(&SubScheduler{}); err != nil {
		log.Printf("基础数据表SubScheduler迁移失败: %v", err)
	} else {
		log.Printf("数据表SubScheduler创建成功")
	}
	if err := db.AutoMigrate(&SystemSetting{}); err != nil {
		log.Printf("基础数据表SystemSetting迁移失败: %v", err)
	} else {
		log.Printf("数据表SystemSetting创建成功")
	}
	if err := db.AutoMigrate(&Script{}); err != nil {
		log.Printf("基础数据表Script迁移失败: %v", err)
	} else {
		log.Printf("数据表Script创建成功")
	}
	if err := db.AutoMigrate(&SubcriptionGroup{}); err != nil {
		log.Printf("基础数据表SubcriptionGroup迁移失败: %v", err)
	} else {
		log.Printf("数据表SubcriptionGroup创建成功")
	}
	if err := db.AutoMigrate(&SubcriptionNode{}); err != nil {
		log.Printf("基础数据表SubcriptionNode迁移失败: %v", err)
	} else {
		log.Printf("数据表SubcriptionNode创建成功")
	}
	if err := db.AutoMigrate(&SubcriptionScript{}); err != nil {
		log.Printf("基础数据表SubcriptionScript迁移失败: %v", err)
	} else {
		log.Printf("数据表SubcriptionScript创建成功")
	}
	if err := db.AutoMigrate(&Template{}); err != nil {
		log.Printf("基础数据表Template迁移失败: %v", err)
	} else {
		log.Printf("数据表Template创建成功")
	}
	if err := db.AutoMigrate(&Tag{}); err != nil {
		log.Printf("基础数据表Tag迁移失败: %v", err)
	} else {
		log.Printf("数据表Tag创建成功")
	}
	if err := db.AutoMigrate(&TagRule{}); err != nil {
		log.Printf("基础数据表TagRule迁移失败: %v", err)
	} else {
		log.Printf("数据表TagRule创建成功")
	}

	// 检查并删除 idx_name_id 索引
	// 0000_drop_idx_name_id
	if err := database.RunCustomMigration("0000_drop_idx_name_id", func() error {
		if db.Migrator().HasIndex(&Node{}, "idx_name_id") {
			if err := db.Migrator().DropIndex(&Node{}, "idx_name_id"); err != nil {
				log.Printf("删除索引 idx_name_id 失败: %v", err)
				return err
			} else {
				log.Println("成功删除索引 idx_name_id")
			}
		}
		return nil
	}); err != nil {
		log.Printf("执行迁移 0000_drop_idx_name_id 失败: %v", err)
	}

	// 0008_node_created_at_fill - 补全空的 CreatedAt 字段
	if err := database.RunCustomMigration("0008_node_created_at_fill", func() error {
		// 查找所有 CreatedAt 为零值的节点并设置为当前时间
		result := db.Exec("UPDATE nodes SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL OR created_at = '' OR created_at = '0001-01-01 00:00:00+00:00'")
		if result.Error != nil {
			return result.Error
		}
		log.Printf("已补全 %d 个节点的创建时间", result.RowsAffected)
		return nil
	}); err != nil {
		log.Printf("执行迁移 0008_node_created_at_fill 失败: %v", err)
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
				log.Printf("Failed to hash password for user %s: %v", user.Username, err)
				continue
			}
			user.Password = hashedPassword
			if err := db.Save(&user).Error; err != nil {
				log.Printf("Failed to save hashed password for user %s: %v", user.Username, err)
			} else {
				log.Printf("Upgraded password for user %s", user.Username)
			}
		}
		return nil
	}); err != nil {
		log.Printf("执行迁移 0005_hash_passwords 失败: %v", err)
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
				log.Printf("增加脚本demo失败: %v", err)
			}
		}
		return nil
	}); err != nil {
		log.Printf("执行迁移 0007_add_script_demo 失败: %v", err)
	}

	// 0009_migrate_template_files - 迁移现有模板文件到数据库
	if err := database.RunCustomMigration("0009_migrate_template_files", func() error {
		return MigrateTemplatesFromFiles("./template")
	}); err != nil {
		log.Printf("执行迁移 0009_migrate_template_files 失败: %v", err)
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
			log.Printf("插入 Clash 基础模板失败: %v", err)
			return err
		}
		// 插入 Surge 模板
		if err := SetSetting("base_template_surge", surgeTemplate); err != nil {
			log.Printf("插入 Surge 基础模板失败: %v", err)
			return err
		}
		log.Println("已添加默认 Clash 和 Surge 基础模板")
		return nil
	}); err != nil {
		log.Printf("执行迁移 0010_add_default_base_templates 失败: %v", err)
	}

	// 0011_migrate_speed_test_concurrency - 迁移旧的并发数配置到新的分离配置
	if err := database.RunCustomMigration("0011_migrate_speed_test_concurrency", func() error {
		// 读取旧的 speed_test_concurrency 配置
		oldConcurrency, _ := GetSetting("speed_test_concurrency")
		if oldConcurrency != "" {
			// 将旧配置迁移到 latency_concurrency
			if err := SetSetting("speed_test_latency_concurrency", oldConcurrency); err != nil {
				log.Printf("迁移 latency_concurrency 失败: %v", err)
				return err
			}
			log.Printf("已将 speed_test_concurrency=%s 迁移到 speed_test_latency_concurrency", oldConcurrency)
		}

		// 设置默认的 speed_concurrency 为 1（如果不存在）
		existingSpeedConcurrency, _ := GetSetting("speed_test_speed_concurrency")
		if existingSpeedConcurrency == "" {
			if err := SetSetting("speed_test_speed_concurrency", "1"); err != nil {
				log.Printf("设置默认 speed_concurrency 失败: %v", err)
				return err
			}
			log.Println("已设置默认 speed_test_speed_concurrency=1")
		}

		// 设置默认的 latency_samples 为 3（如果不存在）
		existingLatencySamples, _ := GetSetting("speed_test_latency_samples")
		if existingLatencySamples == "" {
			if err := SetSetting("speed_test_latency_samples", "3"); err != nil {
				log.Printf("设置默认 latency_samples 失败: %v", err)
				return err
			}
			log.Println("已设置默认 speed_test_latency_samples=3")
		}

		return nil
	}); err != nil {
		log.Printf("执行迁移 0011_migrate_speed_test_concurrency 失败: %v", err)
	}

	// 0012_migrate_last_check_to_separate_fields - 将 LastCheck 字段迁移到 LatencyCheckAt 和 SpeedCheckAt
	if err := database.RunCustomMigration("0012_migrate_last_check_to_separate_fields", func() error {
		// 检查 last_check 列是否存在
		if db.Migrator().HasColumn(&Node{}, "last_check") {
			// 将 last_check 数据复制到 latency_check_at 和 speed_check_at
			result := db.Exec("UPDATE nodes SET latency_check_at = last_check, speed_check_at = last_check WHERE last_check IS NOT NULL AND last_check != ''")
			if result.Error != nil {
				log.Printf("迁移 last_check 数据失败: %v", result.Error)
				return result.Error
			}
			log.Printf("已将 %d 条 last_check 数据迁移到新字段", result.RowsAffected)

			// 删除 last_check 列
			if err := db.Exec("ALTER TABLE nodes DROP COLUMN last_check").Error; err != nil {
				log.Printf("删除 last_check 列失败: %v", err)
				// 不返回错误，因为某些数据库可能不支持 DROP COLUMN
			} else {
				log.Println("成功删除 last_check 列")
			}
		}
		return nil
	}); err != nil {
		log.Printf("执行迁移 0012_migrate_last_check_to_separate_fields 失败: %v", err)
	}

	// 0013_migrate_node_status_fields - 根据已有数据设置 SpeedStatus 和 DelayStatus 字段
	if err := database.RunCustomMigration("0013_migrate_node_status_fields", func() error {
		// DelayTime > 0 且有记录 => DelayStatus = 'success'
		if result := db.Exec("UPDATE nodes SET delay_status = 'success' WHERE delay_time > 0 AND (delay_status IS NULL OR delay_status = '' OR delay_status = 'untested')"); result.Error != nil {
			log.Printf("迁移 DelayStatus (success) 失败: %v", result.Error)
		} else {
			log.Printf("已设置 %d 个节点 DelayStatus 为 success", result.RowsAffected)
		}

		// DelayTime = -1 => DelayStatus = 'timeout'
		if result := db.Exec("UPDATE nodes SET delay_status = 'timeout' WHERE delay_time = -1 AND (delay_status IS NULL OR delay_status = '' OR delay_status = 'untested')"); result.Error != nil {
			log.Printf("迁移 DelayStatus (timeout) 失败: %v", result.Error)
		} else {
			log.Printf("已设置 %d 个节点 DelayStatus 为 timeout", result.RowsAffected)
		}

		// Speed > 0 => SpeedStatus = 'success'
		if result := db.Exec("UPDATE nodes SET speed_status = 'success' WHERE speed > 0 AND (speed_status IS NULL OR speed_status = '' OR speed_status = 'untested')"); result.Error != nil {
			log.Printf("迁移 SpeedStatus (success) 失败: %v", result.Error)
		} else {
			log.Printf("已设置 %d 个节点 SpeedStatus 为 success", result.RowsAffected)
		}

		// Speed = -1 => SpeedStatus = 'error'
		if result := db.Exec("UPDATE nodes SET speed_status = 'error' WHERE speed = -1 AND (speed_status IS NULL OR speed_status = '' OR speed_status = 'untested')"); result.Error != nil {
			log.Printf("迁移 SpeedStatus (error) 失败: %v", result.Error)
		} else {
			log.Printf("已设置 %d 个节点 SpeedStatus 为 error", result.RowsAffected)
		}

		// 所有其他情况 => 'untested'
		if result := db.Exec("UPDATE nodes SET speed_status = 'untested' WHERE speed_status IS NULL OR speed_status = ''"); result.Error != nil {
			log.Printf("迁移 SpeedStatus (untested) 失败: %v", result.Error)
		}
		if result := db.Exec("UPDATE nodes SET delay_status = 'untested' WHERE delay_status IS NULL OR delay_status = ''"); result.Error != nil {
			log.Printf("迁移 DelayStatus (untested) 失败: %v", result.Error)
		}

		log.Println("节点状态字段迁移完成")
		return nil
	}); err != nil {
		log.Printf("执行迁移 0013_migrate_node_status_fields 失败: %v", err)
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
			log.Println("初始化添加用户数据失败")
		}
	} else {
		// Check if we need to update admin password from env
		if envPass := os.Getenv("SUBLINK_ADMIN_PASSWORD_REST"); envPass != "" {
			var admin User
			if err := db.First(&admin).Error; err == nil {
				// Update admin password
				updateUser := &User{Password: envPass}
				if err := admin.Set(updateUser); err != nil {
					log.Printf("Failed to update admin password from env: %v", err)
				} else {
					log.Println("Admin password updated from environment variable")
				}
			}
		}
	}

	// 设置初始化标志为 true
	database.IsInitialized = true
	log.Println("数据库初始化成功")
}
