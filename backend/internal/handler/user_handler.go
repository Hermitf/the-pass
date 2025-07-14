package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/service"
	"github.com/Hermitf/the-pass/internal/utils/userutils"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) LoginHandler(c *gin.Context) {
	// 处理登录逻辑
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 直接传递字段值
    token, err := h.userService.LoginUser(loginReq.LoginInfo, loginReq.Password, loginReq.LoginType)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) RegisterHandler(c *gin.Context) {
	// 处理注册逻辑
	var registerReq RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}
	// 创建用户模型
	passwordHash, err := userutils.GeneratePasswordHash(registerReq.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	registerUser := &model.User{
		Username:     registerReq.Username,
		Email:        registerReq.Email,
		PasswordHash: passwordHash,
		Phone:        registerReq.Phone,
	}
	// 调用用户服务进行注册
	if err := h.userService.RegisterUser(registerUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "registration successful"})
}