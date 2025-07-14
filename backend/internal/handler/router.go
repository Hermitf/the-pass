package handler

import (
	"github.com/Hermitf/the-pass/internal/repository"
	"github.com/Hermitf/the-pass/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 创建路由
func CreateRouter() *gin.Engine {
	router := gin.Default()
	
    // 临时创建服务实例 - 实际应该注入数据库连接
    var db *gorm.DB // 这里需要实际的数据库连接
	userRepo := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo)
    jwtService := service.NewJWTService()
    
    // 添加路由
    AddPublicRoutes(router, userService, jwtService)
    
	return router
}

// 添加公开路由
func AddPublicRoutes(router *gin.Engine, userService *service.UserService, jwtService *service.JWTService) {
    // 用户路由
    userHandler := NewUserHandler(userService)
    router.POST("/users/register", userHandler.RegisterHandler)
    router.POST("/users/login", userHandler.LoginHandler)
    
    // 认证路由
    authHandler := NewJWTHandler(jwtService)
    router.POST("/auth/refresh", authHandler.RefreshToken)
    router.POST("/auth/verify", authHandler.VerifyToken)
}