package middleware

import (
	"net/http"
	"strings"

	"github.com/Hermitf/the-pass/pkg/auth"
	"github.com/gin-gonic/gin"
)

// #region 中间件结构

// JWTMiddleware JWT认证中间件结构体
type JWTMiddleware struct {
	config auth.JWTConfig
}

// NewJWTMiddleware 创建JWT中间件实例
func NewJWTMiddleware(config auth.JWTConfig) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
	}
}

// #endregion

// #region Token提取与验证

// extractAuthHeader 提取Authorization头部
func (m *JWTMiddleware) extractAuthHeader(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供Token"})
		c.Abort()
		return "", false
	}
	return authHeader, true
}

// validateBearerFormat 验证Bearer格式
func (m *JWTMiddleware) validateBearerFormat(c *gin.Context, authHeader string) (string, bool) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token格式错误"})
		c.Abort()
		return "", false
	}

	// 提取token，跳过"Bearer "前缀（7个字符）
	token := strings.TrimSpace(authHeader[7:])
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token不能为空"})
		c.Abort()
		return "", false
	}

	return token, true
}

// verifyTokenAndExtractUserID 验证Token并提取用户ID
func (m *JWTMiddleware) verifyTokenAndExtractUserID(c *gin.Context, token string) (int64, bool) {
	claims, err := auth.VerifyToken(token, m.config)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token无效"})
		c.Abort()
		return 0, false
	}
	return claims.UserID, true
}

// #endregion

// #region 中间件主函数

// AuthMiddleware JWT认证中间件主函数
func (m *JWTMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 步骤1：提取Authorization头部
		authHeader, ok := m.extractAuthHeader(c)
		if !ok {
			return
		}

		// 步骤2：验证Bearer格式并提取token
		token, ok := m.validateBearerFormat(c, authHeader)
		if !ok {
			return
		}

		// 步骤3：验证token并提取用户ID
		userID, ok := m.verifyTokenAndExtractUserID(c, token)
		if !ok {
			return
		}

		// 步骤4：将用户ID存储到上下文中，供后续处理器使用
		c.Set("userID", userID)
		c.Next()
	}
}

// #endregion
