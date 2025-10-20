package config

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Configuration struct {
	Server   ServerConfig   `mapstructure:"server" json:"server" yaml:"server"`
	Database DatabaseConfig `mapstructure:"database" json:"database" yaml:"database"`
	JWT      JWTConfig      `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis" json:"redis" yaml:"redis"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods" json:"allowed_methods" yaml:"allowed_methods"`
}

type ServerConfig struct {
	Port int        `mapstructure:"port" json:"port" yaml:"port"`
	CORS CORSConfig `mapstructure:"cors" json:"cors" yaml:"cors"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DbName   string `mapstructure:"db_name" json:"db_name" yaml:"db_name"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	ExpiresIn int64  `mapstructure:"expires_in" json:"expires_in" yaml:"expires_in"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         int    `mapstructure:"port" json:"port" yaml:"port"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	Database     int    `mapstructure:"database" json:"database" yaml:"database"`
	PoolSize     int    `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns" json:"min_idle_conns" yaml:"min_idle_conns"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	viper  *viper.Viper
	config *Configuration
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		viper:  viper.New(),
		config: &Configuration{},
	}
}

// Load 加载配置
func (cm *ConfigManager) Load(configPath string) error {
	// 设置配置文件路径
	cm.viper.SetConfigFile(configPath)

	// 设置环境变量前缀
	cm.viper.SetEnvPrefix("the_pass")
	cm.viper.AutomaticEnv()

	// 读取配置文件
	if err := cm.viper.ReadInConfig(); err != nil {
		return err
	}

	// 解析配置文件到结构体
	if err := cm.viper.Unmarshal(cm.config); err != nil {
		return err
	}

	log.Println("配置成功加载：", cm.viper.ConfigFileUsed())
	return nil
}

// Watch 监听配置文件变化
func (cm *ConfigManager) Watch() {
	cm.viper.WatchConfig()
	cm.viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("配置文件改变:", e.Name)

		// 重新加载配置
		if err := cm.viper.Unmarshal(cm.config); err != nil {
			log.Println("重新加载配置失败:", err)
		} else {
			log.Println("配置重新加载成功")
		}
	})
	log.Println("配置文件监视器已启动")
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *Configuration {
	return cm.config
}
