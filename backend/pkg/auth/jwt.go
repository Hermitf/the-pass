package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig JWT配置结构
type JWTConfig struct {
	SecretKey string
	ExpiresIn int64
}

// Claims JWT声明结构
type Claims struct {
	UserID   int64  `json:"user_id"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID int64, userType string, jwtConfig JWTConfig) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("用户ID无效")
	}

	if userType == "" {
		return "", fmt.Errorf("用户类型不能为空")
	}

	if jwtConfig.SecretKey == "" {
		return "", fmt.Errorf("JWT密钥未配置")
	}

	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtConfig.ExpiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.SecretKey))
}

// VerifyToken 验证JWT令牌
func VerifyToken(tokenString string, jwtConfig JWTConfig) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("令牌不能为空")
	}

	if jwtConfig.SecretKey == "" {
		return nil, fmt.Errorf("JWT密钥未配置")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法")
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, fmt.Errorf("令牌已过期")
		}
		return nil, fmt.Errorf("令牌无效")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("令牌声明无效")
}
