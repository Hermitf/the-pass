package middleware

import (
	"net/http"
	"strings"

	"github.com/Hermitf/the-pass/internal/service"
	// "github.com/Hermitf/the-pass/internal/utils/userutils"
	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供Token"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token格式错误"})
			c.Abort()
			return
		}

		// extract the token from the header, ignoring the "Bearer " prefix
		token := authHeader[7:]          // 直接截取，跳过"Bearer "（7个字符）
		token = strings.TrimSpace(token) // 移除可能的空格

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token不能为空"})
			c.Abort()
			return
		}

		// 使用jwtService验证token
		userID, err := jwtService.VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token无效"})
			c.Abort()
			return
		}

		// set the user ID in the context for later use
		c.Set("userID", userID)
		c.Next()
	}
}
