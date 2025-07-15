// 项目启动文件
package main

import (
	"log"

	"github.com/Hermitf/the-pass/global"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {

	// 创建默认路由
	router := gin.Default()

	// 测试路由
	// router.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "pong",
	// 	})
	// })
	// 注册路由
	router.GET("/", mainHandler)
	router.POST("/register", registerHandler)

	return router
}

func main() {
	// 初始化配置
	if err := global.App.InitConfig(); err != nil {
		log.Fatal("配置初始化失败:", err)
	}

	// 其他初始化操作
	// server := global.App.Configs.Server

	r := setupRouter()
	// 启动服务
	r.Run(":8080")
}

func registerHandler(c *gin.Context) {
	// 处理注册逻辑
	// w.Write([]byte("Register Handler"))
}

func mainHandler(c *gin.Context) {
	// 处理主页逻辑
	// w.Write([]byte("Main Handler"))
}
