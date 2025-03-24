package models

import (
	"log"
	"os"

	"github.com/glebarez/sqlite" // 替换为纯Go实现的SQLite驱动
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 全局数据库连接
var DB *gorm.DB

// ConnectDatabase 初始化数据库连接
func ConnectDatabase() {
	// 获取数据库路径，默认为data.db
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db"
	}

	// 设置日志级别
	logLevel := logger.Info
	if os.Getenv("GIN_MODE") == "release" {
		logLevel = logger.Error
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	DB = db

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(&User{}, &ChatHistory{})
	if err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}

	log.Println("数据库连接成功")
}