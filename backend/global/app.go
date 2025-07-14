package global

import (
	"log"

	"github.com/jinzhu/gorm"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/Hermitf/the-pass/configs"
)

type Application struct {
	ConfigViper *viper.Viper
	Configs     configs.Configuration
	DB          *gorm.DB
}

func (a *Application) InitConfig() error {
	config := "configs/config.yaml"

	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		log.Fatal("读取配置文件失败: ", err)
		return err
	}

	// 监听配置文件变化
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Println("配置文件已修改:", e.Name)
		// 重新解析配置文件
		if err := v.Unmarshal(&a.Configs); err != nil {
			log.Println("重新解析配置文件失败:", err)
		} else {
			log.Println("配置文件重新加载成功")
		}
	})

	// 解析配置文件到结构体
	if err := v.Unmarshal(&a.Configs); err != nil {
		log.Fatal("解析配置文件失败: ", err)
		return err
	}
	
    a.ConfigViper = v
    return nil
}

var App = new(Application)