package userutils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/Hermitf/the-pass/global"
	"github.com/golang-jwt/jwt/v5"
)

// JWT配置结构
type JWTConfig struct {
	SecretKey string
	ExpiresIn int64 // 过期时间（秒）
}

// getJWTSecret 获取JWT密钥
func getJWTSecret() string {
	// 优先从环境变量获取
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}

	// 返回默认密钥（实际项目中应该从配置文件读取）
	return "K7gNU3sdo+OL0wNhqoVWhr3g6s1xYv72ol/pe/Unols="
}

// GenerateJWTTokenWithConfig 使用自定义配置生成JWT令牌
func GenerateJWTTokenWithConfig(userID int64, config *JWTConfig) (string, error) {
	if config == nil {
		return "", errors.New("the JWT configuration cannot be nil")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": strconv.FormatInt(userID, 10),
		"exp":     now.Add(time.Duration(config.ExpiresIn) * time.Second).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"jti":     strconv.FormatInt(now.UnixNano(), 10), // JWT ID (纳秒时间戳确保唯一性)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.SecretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// VerifyJWTTokenWithConfig 使用自定义配置验证JWT令牌
func VerifyJWTTokenWithConfig(tokenString string, config *JWTConfig) (int64, error) {
	if config == nil {
		return 0, errors.New("the JWT configuration cannot be nil")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		// expired token error handling
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, errors.New("token is expired")
		}
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userIDStr, ok := claims["user_id"].(string); ok {
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				return 0, errors.New("invalid user ID format")
			}
			return userID, nil
		}
	}

	return 0, errors.New("invalid token")
}

// GetJWTClaims 获取JWT令牌中的所有声明
func GetJWTClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(getJWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

func GetJWTConfiguration() *JWTConfig {
	jwtConfig := global.App.Configs.JWT

	config := &JWTConfig{
		SecretKey: jwtConfig.SecretKey,
		ExpiresIn: jwtConfig.ExpiresIn,
	}
	return config
}
