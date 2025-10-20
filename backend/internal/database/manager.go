package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Hermitf/the-pass/internal/config"
	"github.com/Hermitf/the-pass/internal/model"
)

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	db *gorm.DB
}

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{}
}

// Initialize 初始化数据库连接
func (dm *DatabaseManager) Initialize(dbConfig config.DatabaseConfig) error {
	// 检查必需的数据库配置字段
	if dbConfig.Username == "" || dbConfig.Password == "" ||
		dbConfig.Host == "" || dbConfig.Port == 0 || dbConfig.DbName == "" {
		return fmt.Errorf("数据库配置不完整")
	}

	// 构建DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		dbConfig.Host, dbConfig.Username, dbConfig.Password, dbConfig.DbName, dbConfig.Port)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	dm.db = db
	log.Println("数据库连接成功")

	// 自动迁移表结构
	if err := dm.AutoMigrate(); err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}

	return nil
}

// AutoMigrate 自动迁移数据库表结构
func (dm *DatabaseManager) AutoMigrate() error {
	err := dm.db.AutoMigrate(
		&model.User{},
		&model.Employee{},
		&model.Merchant{},
		&model.Rider{},
	)

	if err != nil {
		return err
	}

	log.Println("数据库自动迁移完成")
	return nil
}

// GetDB 获取数据库连接
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

// Close 关闭数据库连接
func (dm *DatabaseManager) Close() error {
	if dm.db != nil {
		sqlDB, err := dm.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
