package handler

import (
	"net/http"

	"github.com/Hermitf/the-pass/internal/service"
	"github.com/gin-gonic/gin"
)

// JWT处理器
type JWTHandler struct {
	jwtService *service.JWTService
}

// 创建JWT处理器
func NewJWTHandler(jwtService *service.JWTService) *JWTHandler {
	return &JWTHandler{jwtService: jwtService}
}

// 刷新Token
func (h *JWTHandler) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		Token string `json:"token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	newToken, err := h.jwtService.RefreshToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// 验证Token
func (h *JWTHandler) VerifyToken(c *gin.Context) {
	type VerifyRequest struct {
		Token string `json:"token" binding:"required"`
	}

	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	userID, err := h.jwtService.VerifyToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"user_id": userID,
	})
}
