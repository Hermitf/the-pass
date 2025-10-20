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

	"github.com/Hermitf/the-pass/internal/model"
)

// configs 类型定义 - 这应该从实际的配置文件中导入
type configs struct {
	Database DatabaseConfig `json:"database" yaml:"database"`
	Redis    RedisConfig    `json:"redis" yaml:"redis"`
	Server   ServerConfig   `json:"server" yaml:"server"`
	JWT      JWTConfig      `json:"jwt" yaml:"jwt"`
}

type DatabaseConfig struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	DbName   string `json:"dbname" yaml:"dbname"`
}

type RedisConfig struct {
	Host         string `json:"host" yaml:"host"`
	Port         int    `json:"port" yaml:"port"`
	Password     string `json:"password" yaml:"password"`
	Database     int    `json:"database" yaml:"database"`
	PoolSize     int    `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns" yaml:"min_idle_conns"`
}

type ServerConfig struct {
	Port int        `json:"port" yaml:"port"`
	CORS CORSConfig `json:"cors" yaml:"cors"`
}

type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods"`
}

type JWTConfig struct {
	SecretKey string `json:"secret_key" yaml:"secret_key"`
	ExpiresIn int64  `json:"expires_in" yaml:"expires_in"`
}

type Application struct {
	ConfigViper *viper.Viper
	Configs     configs
	DB          *gorm.DB
	RedisClient *redis.Client
}

// 初始化配置
func (a *Application) InitConfig() error {
	// 初始化Viper
	a.initViper()

	// 解析配置文件到结构体
	if err := a.ConfigViper.Unmarshal(&a.Configs); err != nil {
		log.Printf("配置文件解析失败: %v", err)
		return err
	}

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
		return fmt.Errorf("Redis连接失败: %v", err)
	}

	log.Println("✅ Redis连接成功")
	return nil
}

// 初始化viper
func (a *Application) initViper() error {
	v := viper.New()

	// 设置配置文件路径和名称
	v.SetConfigFile("./config.yaml")
	// v.SetConfigName("config")

	// 设置环境变量前缀
	v.SetEnvPrefix("the_pass")
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		log.Fatal("配置文件读取失败: ", err)
		return err
	}
	a.ConfigViper = v
	log.Println("配置文件读取成功:", v.ConfigFileUsed())
	return nil
}

// Watch for configuration changes
func (a *Application) watchConfig() {
	a.ConfigViper.WatchConfig()
	a.ConfigViper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("配置文件发生变化:", e.Name)

		// Reload configuration
		if err := a.ConfigViper.Unmarshal(&a.Configs); err != nil {
			log.Println("配置文件重新加载失败:", err)
		} else {
			log.Println("配置文件重新加载成功")
		}
	})
	log.Println("配置文件监控启动")
}

var App = new(Application)
