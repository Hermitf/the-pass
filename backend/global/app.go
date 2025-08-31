package global

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Hermitf/the-pass/configs"
	"github.com/Hermitf/the-pass/internal/model"
)

type Application struct {
	ConfigViper *viper.Viper
	Configs     configs.Configuration
	DB          *gorm.DB
}

// 初始化配置
func (a *Application) InitConfig() error {
	// 初始化Viper
	a.initViper()

	// 解析配置文件到结构体
	if err := a.ConfigViper.Unmarshal(&a.Configs); err != nil {
		log.Printf("Failed to parse config: %v", err)
		return err
	}

	// 监听配置文件变化
	a.watchConfig()

	log.Println("Config initialized successfully:", a.Configs)
	return nil
}

// 初始化数据库
func (a *Application) InitDatabase() error {
	// Check required database config fields
	dbConfig := a.Configs.Database
	if dbConfig.Username == "" || dbConfig.Password == "" || dbConfig.Host == "" || dbConfig.Port == 0 || dbConfig.DbName == "" {
		log.Fatal("Database configuration is not set properly")
		return nil
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		dbConfig.Host, dbConfig.Username, dbConfig.Password, dbConfig.DbName, dbConfig.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil
	}
	a.DB = db
	log.Println("Database connected successfully")

	// create tables and initialize data
	if err := a.DB.AutoMigrate(&model.User{}); err != nil {
		log.Printf("Failed to auto migrate database: %v", err)
		return nil
	}
	log.Println("Database auto migration completed successfully")
	return nil
}

// 初始化viper
func (a *Application) initViper() error {
	v := viper.New()

	// 设置配置文件路径和名称
	v.SetConfigFile("./configs/config.yaml")
	// v.SetConfigName("config")

	// 设置环境变量前缀
	v.SetEnvPrefix("the_pass")
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		log.Fatal("Failed to read config: ", err)
		return err
	}
	a.ConfigViper = v
	log.Println("Config file read successfully:", v.ConfigFileUsed())
	return nil
}

// Watch for configuration changes
func (a *Application) watchConfig() {
	a.ConfigViper.WatchConfig()
	a.ConfigViper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)

		// Reload configuration
		if err := a.ConfigViper.Unmarshal(&a.Configs); err != nil {
			log.Println("Failed to reload config:", err)
		} else {
			log.Println("Config file reloaded successfully")
		}
	})
	log.Println("Config file watcher started")
}

var App = new(Application)
