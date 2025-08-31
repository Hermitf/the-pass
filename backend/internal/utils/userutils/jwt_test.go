package userutils

import (
	"testing"
)

// 测试自定义配置
func TestJWTTokenWithConfig(t *testing.T) {
	userID := int64(456)
	config := &JWTConfig{
		SecretKey: "test-secret-key",
		ExpiresIn: 3600, // 1小时
	}

	token, err := GenerateJWTTokenWithConfig(userID, config)
	if err != nil {
		t.Fatalf("GenerateJWTTokenWithConfig() error: %v", err)
	}

	verifiedUserID, err := VerifyJWTTokenWithConfig(token, config)
	if err != nil {
		t.Errorf("VerifyJWTTokenWithConfig() error: %v", err)
	}

	if verifiedUserID != userID {
		t.Errorf("VerifyJWTTokenWithConfig() userID = %d; want %d", verifiedUserID, userID)
	}
}
