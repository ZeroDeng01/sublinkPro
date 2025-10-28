package models

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var isInitialized bool

// migrateSubcriptionNodeTable 迁移 subcription_nodes 表，将 node_id 转换为 node_name
func migrateSubcriptionNodeTable(db *gorm.DB) {
	// 检查表是否存在
	if !db.Migrator().HasTable("subcription_nodes") {
		log.Println("subcription_nodes 表不存在，跳过迁移")
		return
	}

	// 检查是否存在 node_id 列
	hasNodeID := db.Migrator().HasColumn(&SubcriptionNode{}, "node_id")
	hasNodeName := db.Migrator().HasColumn(&SubcriptionNode{}, "node_name")

	if !hasNodeID {
		log.Println("subcription_nodes 表已迁移（不存在 node_id 列），跳过迁移")
		return
	}

	log.Println("开始迁移 subcription_nodes 表：将 node_id 转换为 node_name")

	// 1. 如果 node_name 列不存在，先创建它
	if !hasNodeName {
		if err := db.Migrator().AddColumn(&SubcriptionNode{}, "node_name"); err != nil {
			log.Printf("添加 node_name 列失败: %v", err)
			return
		}
		log.Println("成功添加 node_name 列")
	}

	// 2. 将 node_id 转换为 node_name（通过 JOIN nodes 表获取 name）
	result := db.Exec(`
		UPDATE subcription_nodes
		SET node_name = (
			SELECT name FROM nodes WHERE nodes.id = subcription_nodes.node_id
		)
		WHERE node_id IS NOT NULL
		AND EXISTS (SELECT 1 FROM nodes WHERE nodes.id = subcription_nodes.node_id)
	`)

	if result.Error != nil {
		log.Printf("更新 node_name 失败: %v", result.Error)
		return
	}

	log.Printf("成功转换 %d 条记录的 node_id 为 node_name", result.RowsAffected)

	// 3. 删除 node_name 为空的记录（这些是找不到对应节点的无效记录）
	deleteResult := db.Exec(`DELETE FROM subcription_nodes WHERE node_name IS NULL OR node_name = ''`)
	if deleteResult.Error != nil {
		log.Printf("删除无效记录失败: %v", deleteResult.Error)
	} else if deleteResult.RowsAffected > 0 {
		log.Printf("删除了 %d 条无效记录（node_name 为空）", deleteResult.RowsAffected)
	}

	// 4. 重建表（SQLite 不支持直接删除主键列，需要重建表）
	// 创建新表
	createNewTableSQL := `
		CREATE TABLE IF NOT EXISTS subcription_nodes_new (
			subcription_id INTEGER NOT NULL,
			node_name TEXT NOT NULL,
			sort INTEGER DEFAULT 0,
			PRIMARY KEY (subcription_id, node_name)
		)
	`
	if err := db.Exec(createNewTableSQL).Error; err != nil {
		log.Printf("创建新表失败: %v", err)
		return
	}
	log.Println("成功创建新表 subcription_nodes_new")

	// 复制数据到新表
	copyDataSQL := `
		INSERT INTO subcription_nodes_new (subcription_id, node_name, sort)
		SELECT subcription_id, node_name, sort
		FROM subcription_nodes
		WHERE node_name IS NOT NULL AND node_name != ''
	`
	copyResult := db.Exec(copyDataSQL)
	if copyResult.Error != nil {
		log.Printf("复制数据到新表失败: %v", copyResult.Error)
		// 清理新表
		db.Exec("DROP TABLE IF EXISTS subcription_nodes_new")
		return
	}
	log.Printf("成功复制 %d 条记录到新表", copyResult.RowsAffected)

	// 删除旧表
	if err := db.Exec("DROP TABLE subcription_nodes").Error; err != nil {
		log.Printf("删除旧表失败: %v", err)
		return
	}
	log.Println("成功删除旧表 subcription_nodes")

	// 重命名新表
	if err := db.Exec("ALTER TABLE subcription_nodes_new RENAME TO subcription_nodes").Error; err != nil {
		log.Printf("重命名新表失败: %v", err)
		return
	}
	log.Println("成功重命名新表为 subcription_nodes")

	log.Println("subcription_nodes 表迁移完成")
}

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
	// 执行数据迁移：将 node_id 转换为 node_name
	migrateSubcriptionNodeTable(db)

	err = db.AutoMigrate(&User{}, &Subcription{}, &Node{}, &SubLogs{}, &AccessKey{}, &SubScheduler{}, &SubcriptionNode{}, &Plugin{})
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
