package service

import (
	"errors"
	"time"

	"github.com/Hermitf/the-pass/global"
	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/utils/userutils"
)

// JWT服务
type JWTService struct{}

// 创建JWT服务
func NewJWTService() *JWTService {
	return &JWTService{}
}

func (s *JWTService) GenerateUserToken(user *model.User) (string, error) {
	// 获取JWT配置
	jwtConfig := global.App.Configs.JWT

	config := &userutils.JWTConfig{
		SecretKey: jwtConfig.SecretKey,
		ExpiresIn: jwtConfig.ExpiresIn,
	}

	return userutils.GenerateJWTTokenWithConfig(user.ID, config)

}

// 验证Token并获取用户ID
func (s *JWTService) VerifyToken(tokenString string) (int64, error) {
	// 从全局配置获取JWT配置
	jwtConfig := global.App.Configs.JWT

	config := &userutils.JWTConfig{
		SecretKey: jwtConfig.SecretKey,
		ExpiresIn: jwtConfig.ExpiresIn,
	}

	return userutils.VerifyJWTTokenWithConfig(tokenString, config)
}

// 刷新Token
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	userID, err := s.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查Token是否即将过期（比如剩余时间少于1小时）
	claims, err := userutils.GetJWTClaims(tokenString)
	if err != nil {
		return "", err
	}

	if exp, ok := claims["exp"].(float64); ok {
		remainingTime := int64(exp) - time.Now().Unix()
		if remainingTime > 3600 { // 1小时
			return "", errors.New("Token还未到刷新时间")
		}
	}

	// 生成新Token
	jwtConfig := global.App.Configs.JWT
	config := &userutils.JWTConfig{
		SecretKey: jwtConfig.SecretKey,
		ExpiresIn: jwtConfig.ExpiresIn,
	}

	return userutils.GenerateJWTTokenWithConfig(userID, config)
}

// 检查Token是否过期
func (s *JWTService) IsTokenExpired(tokenString string) bool {
	return userutils.IsJWTTokenExpired(tokenString)
}
