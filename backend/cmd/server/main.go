// 项目启动文件
package main

import (
	"fmt"
	"log"

	_ "github.com/Hermitf/the-pass/docs"
	"github.com/Hermitf/the-pass/global"
	"github.com/Hermitf/the-pass/internal/handler"
)

// @title The Pass API
// @version 1.0
// @description The Pass API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email api_support@the-pass.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:13544
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 使用Bearer Token进行认证，格式: Bearer {token}

func main() {
	// 初始化配置
	if err := global.App.InitConfig(); err != nil {
		log.Fatal("配置初始化失败:", err)
	}
	// 初始化数据库
	if err := global.App.InitDatabase(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	db := global.App.DB
	if db == nil {
		log.Fatal("数据库未初始化")
	}

	// 创建路由
	r := handler.SetupRouter(db)

	// 启动服务
	port := global.App.Configs.Server.Port
	log.Printf("服务正在监听端口: %d", port)
	log.Printf("Swagger文档地址: http://localhost:%d/swagger/index.html", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("服务启动失败:", err)
	}
}
