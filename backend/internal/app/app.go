package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Hermitf/the-pass/internal/config"
	"github.com/Hermitf/the-pass/internal/database"
)

// AppContext 应用上下文，管理核心依赖和资源
// 这不是整个应用程序，而是应用的核心组件集合
// 真正的应用入口在 cmd/server/main.go
// AppContext 的职责：
// 1. 管理配置、数据库、缓存等核心依赖
// 2. 提供依赖注入服务
// 3. 管理资源的生命周期（初始化和清理）
type AppContext struct {
	Config      *config.Configuration
	DB          *gorm.DB
	RedisClient *redis.Client
}

// NewAppContext 创建应用上下文
func NewAppContext() *AppContext {
	return &AppContext{}
}

// Initialize 初始化应用上下文，加载所有依赖
func (ctx *AppContext) Initialize(configPath string) error {
	// 初始化配置管理器
	configManager := config.NewConfigManager()
	if err := configManager.Load(configPath); err != nil {
		return fmt.Errorf("配置加载失败: %w", err)
	}

	// 启动配置文件监听
	configManager.Watch()

	ctx.Config = configManager.GetConfig()
	log.Println("✅ 配置加载成功")

	// 初始化数据库
	dbManager := database.NewDatabaseManager()
	if err := dbManager.Initialize(ctx.Config.Database); err != nil {
		return fmt.Errorf("数据库初始化失败: %w", err)
	}

	ctx.DB = dbManager.GetDB()
	log.Println("✅ 数据库初始化成功")

	// 初始化Redis
	if err := ctx.initRedis(); err != nil {
		return fmt.Errorf("Redis初始化失败: %w", err)
	}

	log.Println("✅ Redis初始化成功")

	log.Println("🎉 应用上下文初始化完成")
	return nil
}

// initRedis 初始化Redis连接
func (ctx *AppContext) initRedis() error {
	redisConfig := ctx.Config.Redis

	// 创建Redis客户端
	ctx.RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.Database,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
	})

	// 测试连接
	ctx_timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ctx.RedisClient.Ping(ctx_timeout).Err(); err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}

	return nil
}

// Close 关闭所有资源
func (ctx *AppContext) Close() error {
	var errors []error

	// 关闭Redis连接
	if ctx.RedisClient != nil {
		if err := ctx.RedisClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("Redis关闭失败: %w", err))
		}
	}

	// 关闭数据库连接
	if ctx.DB != nil {
		sqlDB, err := ctx.DB.DB()
		if err != nil {
			errors = append(errors, fmt.Errorf("获取SQL DB失败: %w", err))
		} else if err := sqlDB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("数据库关闭失败: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("应用上下文关闭时发生错误: %v", errors)
	}

	log.Println("✅ 应用上下文关闭成功")
	return nil
}
