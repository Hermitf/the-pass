package handler

import (
	"gorm.io/gorm"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/Hermitf/the-pass/global"
	"github.com/Hermitf/the-pass/internal/middleware"
	"github.com/Hermitf/the-pass/internal/repository"
	"github.com/Hermitf/the-pass/internal/service"
)

// create a new Gin router and set up the routes
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// 从配置获取CORS设置
	corsConfig := cors.Config{
		AllowOrigins:     global.App.Configs.Server.CORS.AllowedOrigins,
		AllowMethods:     global.App.Configs.Server.CORS.AllowedMethods,
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// add Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// create service instances
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	jwtService := service.NewJWTService()

	// create handler instances
	userHandler := NewUserHandler(userService)

	// set up API routes
	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{

			// User related API
			users := v1.Group("/users")
			{
				// 无需认证的接口
				users.POST("/register", userHandler.RegisterHandler) // POST /api/v1/users/register
				users.POST("/login", userHandler.LoginHandler)       // POST /api/v1/users/login

				// 需要认证的接口
				authenticated := users.Group("")
				authenticated.Use(middleware.JWTAuthMiddleware(jwtService)) // ← 这行很关键
				{
					authenticated.GET("/profile", userHandler.GetProfileHandler)
					authenticated.PUT("/profile", userHandler.UpdateProfileHandler)
				}
				// users.Use(middleware.JWTAuthMiddleware(jwtService)) // Use JWT middleware for authentication
				// users.GET("/profile", userHandler.GetProfileHandler) // GET /api/v1/users/profile
				// users.PUT("/profile", userHandler.UpdateProfileHandler) // PUT /api/v1/users/profile
			}
		}
	}

	return router
}
