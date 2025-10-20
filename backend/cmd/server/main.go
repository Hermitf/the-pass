// 项目启动文件
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hermitf/the-pass/internal/app"
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
	// 创建应用上下文（核心依赖管理）
	appCtx := app.NewAppContext()

	// 初始化应用上下文
	if err := appCtx.Initialize("./config.yaml"); err != nil {
		log.Fatal("应用上下文初始化失败:", err)
	}

	// 设置优雅关闭
	defer func() {
		if err := appCtx.Close(); err != nil {
			log.Printf("关闭应用上下文时出错: %v", err)
		}
	}()

	// 创建路由（传入应用上下文）
	router := handler.NewRouter(appCtx)

	// 启动服务
	port := appCtx.Config.Server.Port
	log.Printf("🚀 服务正在监听端口: %d", port)
	log.Printf("📚 Swagger文档地址: http://localhost:%d/swagger/index.html", port)

	// 创建错误通道
	errCh := make(chan error, 1)

	// 启动HTTP服务器
	go func() {
		if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
			errCh <- fmt.Errorf("服务启动失败: %w", err)
		}
	}()

	// 监听系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待错误或信号
	select {
	case err := <-errCh:
		log.Fatal(err)
	case sig := <-sigCh:
		log.Printf("📍 接收到信号: %v, 正在优雅关闭...", sig)
	}

	log.Println("✅ 服务已关闭")
}
