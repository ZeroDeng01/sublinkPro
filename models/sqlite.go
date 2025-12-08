package models

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var isInitialized bool

func InitSqlite() {
	// 检查目录是否创建
	_, err := os.Stat("./db")
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("./db", os.ModePerm)
		}
	}
	// 连接数据库
	db, err := gorm.Open(sqlite.Open("./db/sublink.db"), &gorm.Config{})
	if err != nil {
		log.Println("连接数据库失败")
	}
	DB = db
	// 检查是否已经初始化
	if isInitialized {
		log.Println("数据库已经初始化，无需重复初始化")
		return
	}

	// 检查并删除 idx_name_id 索引
	// 0000_drop_idx_name_id
	if err := RunCustomMigration("0000_drop_idx_name_id", func() error {
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

	// 0001_initial_tables
	if err := RunAutoMigrate("0001_initial_tables", &User{}, &Subcription{}, &Node{}, &SubLogs{}, &AccessKey{}, &SubScheduler{}, &SystemSetting{}, &Script{}); err != nil {
		log.Printf("基础数据表迁移失败: %v", err)
	}

	// 0001_node_add_country 添加国家字段
	if err := RunAutoMigrate("0001_node_add_country", &Node{}); err != nil {
		log.Printf("Nodes 表迁移失败: %v", err)
	}

	if err := RunAutoMigrate("0002_subcription_node", &SubcriptionNode{}); err != nil {
		log.Printf("SubcriptionNode 表迁移失败: %v", err)
	}

	// SubcriptionScript 单独处理
	// 0003_subcription_script
	if err := RunAutoMigrate("0003_subcription_script", &SubcriptionScript{}); err != nil {
		log.Printf("SubcriptionScript 表迁移失败: %v", err)
	}

	// 创建 SubcriptionGroup 表
	// 0004_subcription_group
	if err := RunAutoMigrate("0004_subcription_group", &SubcriptionGroup{}); err != nil {
		log.Printf("SubcriptionGroup 表迁移失败: %v", err)
	}

	// 0006_subscheduler_proxy
	if err := RunAutoMigrate("0006_subscheduler_proxy", &SubScheduler{}); err != nil {
		log.Printf("SubScheduler 表新增代理字段迁移失败: %v", err)
	}

	// 0007_node_timestamps - 确保 Node 表有 created_at 和 updated_at 列
	if err := RunAutoMigrate("0007_node_add_created_at_and_updated_at", &Node{}); err != nil {
		log.Printf("Node 表时间戳字段迁移失败: %v", err)
	}

	// 0008_node_created_at_fill - 补全空的 CreatedAt 字段
	if err := RunCustomMigration("0008_node_created_at_fill", func() error {
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
	if err := RunCustomMigration("0005_hash_passwords", func() error {
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
	if err := RunCustomMigration("0007_add_script_demo", func() error {
		script := &Script{}
		script.Name = "[系统DEMO]按测速结果筛选节点"
		script.Content = "" +
			"//修改节点列表\n/**\n * @param {Node[]} nodes\n * @param {string} clientType\n */\nfunction filterNode(nodes, clientType) {\n    let maxDelayTime = 250;//最大延迟 单位ms \n    let minSpeed = 1.5;//最小速度 单位MB/s\n    // nodes: 节点列表\n    // 数据结构如下\n    // [\n    //     {\n    //         \"ID\": 1,\n    //         \"Link\": \"vmess://4564564646\",\n    //         \"Name\": \"xx订阅_US-CDN-SSL\",\n    //         \"LinkName\": \"US-CDN-SSL\",\n    //         \"LinkAddress\": \"xxxxxxxxx.net:443\",\n    //         \"LinkHost\": \"xxxxxxxxx.net\",\n    //         \"LinkPort\": \"443\",\n    //         \"DialerProxyName\": \"\",\n    //         \"CreateDate\": \"\",\n    //         \"Source\": \"manual\",\n    //         \"SourceID\": 0,\n    //         \"Group\": \"自用\",\n    //         \"DelayTime\": 110,\n    //         \"Speed\": 10,\n    //         \"LastCheck\": \"2025-11-26 23:49:58\"\n    //     }\n    // ]\n    // clientType: 客户端类型\n    // 返回值: 修改后节点列表\n    let newNodes = [];\n    nodes.forEach(node => {\n        if(!node.Link.includes(\"://_\")){\n            //如果分组是机场或者自用的自建节点则忽略测速直接加入列表\n            if(node.Group.includes(\"机场\")||node.Group.includes(\"自建\")){\n                newNodes.push(rename(node));\n            }else{\n                //速度高或者延迟低都保留\n                if(node.DelayTime>0&&(node.DelayTime<maxDelayTime||node.Speed>=minSpeed)){\n                    newNodes.push(rename(node));\n                    console.log(\"✅节点：\"+node.Name +\"符合测速要求\");\n                }else{\n                    console.log(\"❌节点：\"+node.Name +\"不符合测速要求\");\n                }\n            }\n        }\n    });\n    return newNodes;\n}\n//修改订阅文件\n/**\n * @param {string} input\n * @param {string} clientType\n */\nfunction subMod( input, clientType) {\n    // input: 原始输入内容,不同客户端订阅文件也不一样\n    // clientType: 客户端类型\n    // 返回值: 修改后的内容字符串\n    return input; // 注意：此处示例仅为示意，实际应返回处理后的字符串\n}\n\n// 节点改名\nfunction rename(node){\n    if(node.Link.indexOf('#')!=-1&&node.Source!=='manual'){\n        var linkArr = node.Link.split('#')\n        node.Link = linkArr[0]+'#'+node.Source+\"_\"+linkArr[1]\n        return node\n    }\n\n    return node\n}"
		script.Version = "1.0.0"
		if script.CheckNameVersion() {
			return nil
		}
		err = db.First(&Script{}).Error
		if err == gorm.ErrRecordNotFound {
			err := script.Add()
			if err != nil {
				log.Printf("增加脚本demo失败: %v", err)
			}
		}
		return nil
	}); err != nil {
		log.Printf("执行迁移 0000_drop_idx_name_id 失败: %v", err)
	}

	// 初始化用户数据
	err = db.First(&User{}).Error
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
		if envPass := os.Getenv("SUBLINK_ADMIN_PASSWORD"); envPass != "" {
			var admin User
			if err := db.Where("username = ?", "admin").First(&admin).Error; err == nil {
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
	isInitialized = true
	log.Println("数据库初始化成功") // 只有在没有任何错误时才会打印这个日志
}
