package service

import (
	"github.com/Hermitf/the-pass/pkg/auth"
)

// #region 服务定义

// JWTServiceInterface JWT服务接口
type JWTServiceInterface interface {
	GenerateToken(userID int64, userType string) (string, error)
	VerifyToken(tokenString string) (int64, error)
	RefreshToken(tokenString string) (string, error)
}

// JWTService JWT服务实现
type JWTService struct {
	config auth.JWTConfig
}

// #endregion

// #region 构造函数

// NewJWTService 创建JWT服务实例
func NewJWTService(config auth.JWTConfig) JWTServiceInterface {
	return &JWTService{
		config: config,
	}
}

// #endregion

// #region JWT操作

// GenerateToken 为任意用户类型生成JWT令牌
func (s *JWTService) GenerateToken(userID int64, userType string) (string, error) {
	return auth.GenerateToken(userID, userType, s.config)
}

// VerifyToken 验证JWT令牌并返回用户ID
func (s *JWTService) VerifyToken(tokenString string) (int64, error) {
	claims, err := auth.VerifyToken(tokenString, s.config)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// RefreshToken 刷新JWT令牌
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := auth.VerifyToken(tokenString, s.config)
	if err != nil {
		return "", err
	}

	// 生成新令牌
	return auth.GenerateToken(claims.UserID, claims.UserType, s.config)
}

// #endregion
