package main

import (
	"log"

	"github.com/Hermitf/the-pass/internal/app"
)

func main() {
	// 只测试Redis连接，不启动整个服务
	log.Println("=== Redis连接测试 ===")

	// 创建应用上下文
	appCtx := app.NewAppContext()

	// 初始化应用上下文（包含配置和Redis）
	if err := appCtx.Initialize("./config.yaml"); err != nil {
		log.Fatal("应用上下文初始化失败:", err)
	}
	defer appCtx.Close()

	log.Println("🎉 Redis测试成功！准备开始实际应用开发")
}
