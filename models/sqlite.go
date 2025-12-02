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

	// SubcriptionNode 可能会因为手动迁移逻辑导致问题，单独处理
	// 0002_subcription_node
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
