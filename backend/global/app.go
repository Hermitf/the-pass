package global

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Hermitf/the-pass/internal/config"
	"github.com/Hermitf/the-pass/internal/model"
)

// Application 全局应用实例
type Application struct {
	Configs     *config.Configuration // 使用导入的配置类型
	ConfigViper *viper.Viper
	DB          *gorm.DB
	RedisClient *redis.Client
}

// 初始化配置
func (a *Application) InitConfig() error {
	if err := a.initViper(); err != nil {
		return err
	}

	cfg := new(config.Configuration)
	if err := a.ConfigViper.Unmarshal(cfg); err != nil {
		log.Printf("配置文件解析失败: %v", err)
		return err
	}
	a.Configs = cfg

	// 监听配置文件变化
	a.watchConfig()

	log.Println("配置初始化成功:", a.Configs)
	return nil
}

// 初始化数据库
func (a *Application) InitDatabase() error {
	// Check required database config fields
	dbConfig := a.Configs.Database
	if dbConfig.Username == "" || dbConfig.Password == "" || dbConfig.Host == "" || dbConfig.Port == 0 || dbConfig.DbName == "" {
		log.Fatal("数据库配置设置不正确")
		return nil
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		dbConfig.Host, dbConfig.Username, dbConfig.Password, dbConfig.DbName, dbConfig.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		return nil
	}
	a.DB = db
	log.Println("数据库连接成功")

	// create tables and initialize data
	if err := a.DB.AutoMigrate(
		&model.User{},
		&model.Employee{},
		&model.Merchant{},
		&model.Rider{},
	); err != nil {
		log.Printf("数据库自动迁移失败: %v", err)
		return err
	}
	log.Println("数据库自动迁移完成")
	return nil
}

// 初始化Redis
func (a *Application) InitRedis() error {
	redisConfig := a.Configs.Redis

	// 直接创建Redis客户端
	a.RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.Database,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.RedisClient.Ping(ctx).Err(); err != nil {
		a.RedisClient.Close()
		a.RedisClient = nil
		return fmt.Errorf("redis连接失败: %w", err)
	}

	log.Println("✅ Redis连接成功")
	return nil
}

// initViper 初始化配置读取器
func (a *Application) initViper() error {
	if a.ConfigViper != nil {
		return nil
	}

	v := viper.New()
	v.SetConfigFile("./config.yaml")
	v.SetEnvPrefix("the_pass")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Printf("配置文件读取失败: %v", err)
		return err
	}

	a.ConfigViper = v
	log.Println("配置文件读取成功:", v.ConfigFileUsed())
	return nil
}

// watchConfig 监听配置文件变化并自动刷新配置
func (a *Application) watchConfig() {
	if a.ConfigViper == nil {
		return
	}

	a.ConfigViper.WatchConfig()
	a.ConfigViper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("配置文件发生变化:", e.Name)

		cfg := new(config.Configuration)
		if err := a.ConfigViper.Unmarshal(cfg); err != nil {
			log.Println("配置文件重新加载失败:", err)
			return
		}
		a.Configs = cfg
		log.Println("配置文件重新加载成功")
	})
	log.Println("配置文件监控启动")
}

var App = new(Application)
