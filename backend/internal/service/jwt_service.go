package service

import (
	"errors"
	"time"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/utils/userutils"
)

// JWT Service
type JWTService struct{}

// create a new JWTService instance
func NewJWTService() *JWTService {
	return &JWTService{}
}

// Generate a JWT token for a user
func (s *JWTService) GenerateUserToken(user *model.User) (string, error) {
	config := userutils.GetJWTConfiguration()

	return userutils.GenerateJWTTokenWithConfig(user.ID, config)

}

// Verify a JWT token and return user ID
func (s *JWTService) VerifyToken(tokenString string) (int64, error) {
	config := userutils.GetJWTConfiguration()
	userID, err := userutils.VerifyJWTTokenWithConfig(tokenString, config)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// Refresh Token
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	config := userutils.GetJWTConfiguration()
	userID, err := userutils.VerifyJWTTokenWithConfig(tokenString, config)
	if err != nil {
		return "", err
	}

	// check if the token is about to expire
	claims, err := userutils.GetJWTClaims(tokenString)
	if err != nil {
		return "", err
	}

	if exp, ok := claims["exp"].(float64); ok {
		remainingTime := int64(exp) - time.Now().Unix()
		if remainingTime > 3600 { // 1小时
			return "", errors.New("token has not expired yet")
		}
	}

	// 生成新Token
	config = userutils.GetJWTConfiguration()

	return userutils.GenerateJWTTokenWithConfig(userID, config)
}
