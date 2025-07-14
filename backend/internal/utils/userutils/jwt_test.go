package userutils

import (
    "strings"
    "testing"
    "time"
)

// 测试JWT Token生成和验证
func TestJWTToken(t *testing.T) {
    userID := int64(123)
    
    // 生成token
    token, err := GenerateJWTToken(userID)
    if err != nil {
        t.Fatalf("GenerateJWTToken() error: %v", err)
    }
    
    if token == "" {
        t.Error("GenerateJWTToken() should return non-empty token")
    }
    
    // 验证token格式
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        t.Errorf("GenerateJWTToken() token format invalid, got %d parts, want 3", len(parts))
    }
    
    // 验证token
    verifiedUserID, err := VerifyJWTToken(token)
    if err != nil {
        t.Errorf("VerifyJWTToken() error: %v", err)
    }
    
    if verifiedUserID != userID {
        t.Errorf("VerifyJWTToken() userID = %d; want %d", verifiedUserID, userID)
    }
}

// 测试JWT Token边界情况
func TestJWTTokenEdgeCases(t *testing.T) {
    testCases := []struct {
        name   string
        userID int64
    }{
        {"positive ID", 123},
        {"zero ID", 0},
        {"negative ID", -1},
        {"large ID", 4294967295},
    }

    for _, tt := range testCases {
        t.Run(tt.name, func(t *testing.T) {
            token, err := GenerateJWTToken(tt.userID)
            if err != nil {
                t.Errorf("GenerateJWTToken(%d) error = %v; want nil", tt.userID, err)
            }
            
            if token == "" {
                t.Errorf("GenerateJWTToken(%d) returned empty token", tt.userID)
            }
            
            verifiedUserID, err := VerifyJWTToken(token)
            if err != nil {
                t.Errorf("VerifyJWTToken() error = %v; want nil", err)
            }
            
            if verifiedUserID != tt.userID {
                t.Errorf("VerifyJWTToken() userID = %d; want %d", verifiedUserID, tt.userID)
            }
        })
    }
}

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

// 测试无效JWT Token
func TestInvalidJWTToken(t *testing.T) {
    invalidTokens := []struct {
        name  string
        token string
    }{
        {"empty token", ""},
        {"invalid format", "invalid.token"},
        {"malformed token", "header.payload"},
        {"random string", "this-is-not-a-jwt-token"},
    }
    
    for _, tt := range invalidTokens {
        t.Run(tt.name, func(t *testing.T) {
            _, err := VerifyJWTToken(tt.token)
            if err == nil {
                t.Errorf("VerifyJWTToken(%q) should return error, got nil", tt.token)
            }
        })
    }
}

// 测试JWT令牌过期检查
func TestJWTTokenExpiration(t *testing.T) {
    userID := int64(789)
    
    config := &JWTConfig{
        SecretKey: "test-secret",
        ExpiresIn: 1, // 1秒后过期
    }
    
    token, err := GenerateJWTTokenWithConfig(userID, config)
    if err != nil {
        t.Fatalf("GenerateJWTTokenWithConfig() error: %v", err)
    }
    
    // 立即检查，应该没有过期（使用自定义配置）
    if IsJWTTokenExpiredWithConfig(token, config) {
        t.Error("Token should not be expired immediately")
    }
    
    // 等待2秒后检查
    time.Sleep(2 * time.Second)

    if !IsJWTTokenExpiredWithConfig(token, config) {
        t.Error("Token should be expired after 2 seconds")
    }
}

// 测试JWT令牌刷新
func TestRefreshJWTToken(t *testing.T) {
    userID := int64(999)
    
    originalToken, err := GenerateJWTToken(userID)
    if err != nil {
        t.Fatalf("GenerateJWTToken() error: %v", err)
    }
    
    // 添加小延迟确保时间戳不同
    time.Sleep(10 * time.Millisecond)
    
    newToken, err := RefreshJWTToken(originalToken)
    if err != nil {
        t.Fatalf("RefreshJWTToken() error: %v", err)
    }
    
    verifiedUserID, err := VerifyJWTToken(newToken)
    if err != nil {
        t.Errorf("VerifyJWTToken() error: %v", err)
    }
    
    if verifiedUserID != userID {
        t.Errorf("RefreshJWTToken() userID = %d; want %d", verifiedUserID, userID)
    }
    
    if newToken == originalToken {
        t.Error("RefreshJWTToken() should generate a different token")
    }
}

// 性能测试
func BenchmarkJWTFunctions(b *testing.B) {
    userID := int64(12345)
    
    b.Run("GenerateJWTToken", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            GenerateJWTToken(userID)
        }
    })
    
    b.Run("VerifyJWTToken", func(b *testing.B) {
        token, _ := GenerateJWTToken(userID)
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            VerifyJWTToken(token)
        }
    })
    
    b.Run("IsJWTTokenExpired", func(b *testing.B) {
        token, _ := GenerateJWTToken(userID)
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            IsJWTTokenExpired(token)
        }
    })
}

// 测试超出安全范围的整数（应该正确处理）
func TestJWTTokenUnsafeLargeNumbers(t *testing.T) {
    // 这些测试现在应该通过，因为我们使用字符串存储
    testCases := []struct {
        name   string
        userID int64
    }{
        {"max int64", 9223372036854775807},
        {"min int64", -9223372036854775808},
    }

    for _, tt := range testCases {
        t.Run(tt.name, func(t *testing.T) {
            token, err := GenerateJWTToken(tt.userID)
            if err != nil {
                t.Errorf("GenerateJWTToken(%d) error = %v; want nil", tt.userID, err)
            }
            
            verifiedUserID, err := VerifyJWTToken(token)
            if err != nil {
                t.Errorf("VerifyJWTToken() error = %v; want nil", err)
            }
            
            if verifiedUserID != tt.userID {
                t.Errorf("VerifyJWTToken() userID = %d; want %d", verifiedUserID, tt.userID)
            }
        })
    }
}

