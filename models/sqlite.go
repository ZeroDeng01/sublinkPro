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

	tables := []interface{}{
		&User{}, &Subcription{}, &Node{}, &SubLogs{}, &AccessKey{},
		&SubScheduler{}, &SystemSetting{}, &Script{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			log.Printf("数据表迁移失败: %T, %v", table, err)
		}
	}

	// SubcriptionNode 可能会因为手动迁移逻辑导致问题，单独处理
	if err := db.AutoMigrate(&SubcriptionNode{}); err != nil {
		log.Printf("SubcriptionNode 表迁移失败: %v", err)
	}

	// SubcriptionScript 单独处理
	if err := db.AutoMigrate(&SubcriptionScript{}); err != nil {
		log.Printf("SubcriptionScript 表迁移失败: %v", err)
	}

	// 创建 SubcriptionGroup 表
	err = db.AutoMigrate(&SubcriptionGroup{})
	if err != nil {
		log.Println("数据表迁移失败")
	}
	// 初始化用户数据
	err = db.First(&User{}).Error
	if err == gorm.ErrRecordNotFound {
		admin := &User{
			Username: "admin",
			Password: "123456",
			Role:     "admin",
			Nickname: "管理员",
		}
		err = admin.Create()
		if err != nil {
			log.Println("初始化添加用户数据失败")
		}
	}
	// 设置初始化标志为 true
	isInitialized = true
	log.Println("数据库初始化成功") // 只有在没有任何错误时才会打印这个日志
}
