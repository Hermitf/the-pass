package handler

import (
	"time"

	"github.com/Hermitf/the-pass/internal/app"
	"github.com/Hermitf/the-pass/internal/middleware"
	"github.com/Hermitf/the-pass/internal/repository"
	"github.com/Hermitf/the-pass/internal/service"
	"github.com/Hermitf/the-pass/pkg/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouterDependencies holds all the dependencies needed for route setup
type RouterDependencies struct {
	AuthHandler     *AuthHandler
	MerchantHandler *MerchantHandler
	RiderHandler    *RiderHandler
	JWTMiddleware   *middleware.JWTMiddleware
}

// setupMiddleware 配置CORS和其他中间件
func setupMiddleware(router *gin.Engine, appCtx *app.AppContext) {
	// CORS middleware configuration
	corsConfig := cors.Config{
		AllowOrigins:     appCtx.Config.Server.CORS.AllowedOrigins,
		AllowMethods:     appCtx.Config.Server.CORS.AllowedMethods,
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(corsConfig))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
}

// initializeDependencies creates and returns all dependencies needed for routing
func initializeDependencies(appCtx *app.AppContext) *RouterDependencies {
	// Initialize repositories
	userRepo := repository.NewUserRepository(appCtx.DB)
	employeeRepo := repository.NewEmployeeRepository(appCtx.DB)
	merchantRepo := repository.NewMerchantRepository(appCtx.DB)
	riderRepo := repository.NewRiderRepository(appCtx.DB)

	// Create JWT config from application context configuration
	jwtConfig := auth.JWTConfig{
		SecretKey: appCtx.Config.JWT.SecretKey,
		ExpiresIn: appCtx.Config.JWT.ExpiresIn,
	}

	// Initialize shared JWT service
	jwtService := service.NewJWTService(jwtConfig)

	// Initialize services with proper dependencies
	userService := service.NewUserService(service.UserServiceDependencies{
		UserRepo:   userRepo,
		JWTService: jwtService,
	})
	employeeService := service.NewEmployeeService(service.EmployeeServiceDependencies{
		EmployeeRepo: employeeRepo,
		JWTService:   jwtService,
	})
	merchantService := service.NewMerchantService(service.MerchantServiceDependencies{
		MerchantRepo: merchantRepo,
		EmployeeRepo: employeeRepo,
		JWTService:   jwtService,
	})
	riderService := service.NewRiderService(service.RiderServiceDependencies{
		RiderRepo:  riderRepo,
		JWTService: jwtService,
	})

	// Initialize handlers
	authHandler := NewAuthHandler(userService, employeeService, merchantService, riderService)
	merchantHandler := NewMerchantHandler(merchantService, employeeService)
	riderHandler := NewRiderHandler(riderService)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(jwtConfig)

	return &RouterDependencies{
		AuthHandler:     authHandler,
		MerchantHandler: merchantHandler,
		RiderHandler:    riderHandler,
		JWTMiddleware:   jwtMiddleware,
	}
}

// setupSwaggerRoutes configures Swagger documentation routes
func setupSwaggerRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// setupPublicRoutes configures all public routes (no authentication required)
func setupPublicRoutes(v1 *gin.RouterGroup, deps *RouterDependencies) (*gin.RouterGroup, *gin.RouterGroup, *gin.RouterGroup, *gin.RouterGroup) {
	// User routes
	userGroup := v1.Group("/users")
	{
		userGroup.POST("/register", deps.AuthHandler.RegisterHandler("user"))
		userGroup.POST("/login", deps.AuthHandler.LoginHandler("user"))
	}

	// Employee routes
	employeeGroup := v1.Group("/employees")
	{
		employeeGroup.POST("/register", deps.AuthHandler.RegisterHandler("employee"))
		employeeGroup.POST("/login", deps.AuthHandler.LoginHandler("employee"))
	}

	// Rider routes
	riderGroup := v1.Group("/riders")
	{
		riderGroup.POST("/register", deps.AuthHandler.RegisterHandler("rider"))
		riderGroup.POST("/login", deps.AuthHandler.LoginHandler("rider"))
	}

	// Merchant routes
	merchantGroup := v1.Group("/merchants")
	{
		merchantGroup.POST("/register", deps.AuthHandler.RegisterHandler("merchant"))
		merchantGroup.POST("/login", deps.AuthHandler.LoginHandler("merchant"))
	}

	return userGroup, employeeGroup, riderGroup, merchantGroup
}

// setupUserProtectedRoutes configures user-specific protected routes
func setupUserProtectedRoutes(userGroup *gin.RouterGroup, deps *RouterDependencies) {
	usersAuth := userGroup.Group("")
	usersAuth.Use(deps.JWTMiddleware.AuthMiddleware())
	{
		usersAuth.GET("/profile", deps.AuthHandler.GetProfileHandler("user"))
	}
}

// setupEmployeeProtectedRoutes configures employee-specific protected routes
func setupEmployeeProtectedRoutes(employeeGroup *gin.RouterGroup, deps *RouterDependencies) {
	employeesAuth := employeeGroup.Group("")
	employeesAuth.Use(deps.JWTMiddleware.AuthMiddleware())
	{
		employeesAuth.GET("/profile", deps.AuthHandler.GetProfileHandler("employee"))
	}
}

// setupRiderProtectedRoutes configures rider-specific protected routes
func setupRiderProtectedRoutes(riderGroup *gin.RouterGroup, deps *RouterDependencies) {
	ridersAuth := riderGroup.Group("")
	ridersAuth.Use(deps.JWTMiddleware.AuthMiddleware())
	{
		// Common routes (unified handler)
		ridersAuth.GET("/profile", deps.AuthHandler.GetProfileHandler("rider"))

		// Rider-specific business routes (specialized handler)
		ridersAuth.PUT("/online-status", deps.RiderHandler.UpdateOnlineStatusHandler)
		ridersAuth.PUT("/location", deps.RiderHandler.UpdateLocationHandler)
	}
}

// setupMerchantProtectedRoutes configures merchant-specific protected routes
func setupMerchantProtectedRoutes(merchantGroup *gin.RouterGroup, deps *RouterDependencies) {
	merchantsAuth := merchantGroup.Group("")
	merchantsAuth.Use(deps.JWTMiddleware.AuthMiddleware())
	{
		// Common routes (unified handler)
		merchantsAuth.GET("/profile", deps.AuthHandler.GetProfileHandler("merchant"))

		// Merchant-specific business routes (specialized handlers)
		merchantsAuth.POST("/employees", deps.AuthHandler.AddEmployeeHandler())
		merchantsAuth.GET("/employees", deps.MerchantHandler.GetEmployeesHandler)
	}
}

// SetupRouter creates a new Gin router and sets up all routes
// NewRouter creates a new router with dependency injection
func NewRouter(appCtx *app.AppContext) *gin.Engine {
	router := gin.Default()

	// Setup middleware
	setupMiddleware(router, appCtx)

	// Initialize all dependencies
	deps := initializeDependencies(appCtx)

	// Setup API group and Swagger
	v1 := router.Group("/api/v1")
	setupSwaggerRoutes(router)

	// Setup public routes
	userGroup, employeeGroup, riderGroup, merchantGroup := setupPublicRoutes(v1, deps)

	// Setup protected routes for each user type
	setupUserProtectedRoutes(userGroup, deps)
	setupEmployeeProtectedRoutes(employeeGroup, deps)
	setupRiderProtectedRoutes(riderGroup, deps)
	setupMerchantProtectedRoutes(merchantGroup, deps)

	return router
}
