package userutils

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// 测试所有密码相关功能
func TestPasswordFunctions(t *testing.T) {
	testCases := []struct {
		name           string
		password       string
		shouldHash     bool
		isStrong       bool
		expectedIssues int
	}{
		// 基本功能测试
		{"correct password", "testpassword123", true, false, 2}, // 缺少大写和特殊字符
		{"strong password", "MyPass123!", true, true, 0},
		{"empty password", "", false, false, 5}, // 长度+大写+小写+数字+特殊
		{"space only", "   ", false, false, 5},  // 长度+大写+小写+数字+特殊

		// 密码强度测试
		{"no uppercase", "mypass123!", true, false, 1},
		{"no lowercase", "MYPASS123!", true, false, 1},
		{"no number", "MyPassword!", true, false, 1},
		{"no special", "MyPassword123", true, false, 1},
		{"too short", "MyP1!", true, false, 1},
		{"exactly 8 chars strong", "MyPass1!", true, true, 0},
		{"long strong", "MyVeryLongPassword123!", true, true, 0},

		// 边界情况
		{"only spaces", "        ", false, false, 4}, // 大写+小写+数字+特殊（8位空格不算短）
		{"only letters", "abcdefgh", true, false, 3},
		{"only numbers", "12345678", true, false, 3},
		{"only special", "!@#$%^&*", true, false, 3},

		// 特殊字符测试
		{"with spaces", "test password", true, false, 3},        // 缺少大写+数字+特殊
		{"with special chars", "test@pass#123", true, false, 1}, // 缺少大写
		{"unicode password", "密码123", true, false, 3},           // 缺少大写+小写+特殊
		{"unicode strong", "Mý密码123!", true, true, 0},
		{"emoji password", "pass🔐word", true, false, 2},
		{"emoji weak", "🔐🔐🔐🔐🔐🔐🔐🔐", true, false, 3},

		// 长度测试（调整为72字节以内）
		{"long password", strings.Repeat("a", 60), true, false, 3},
		{"long strong", "A" + strings.Repeat("a", 55) + "1!", true, true, 0},
		{"exactly 72 chars", "A" + strings.Repeat("a", 68) + "1!", true, true, 0},
		{"too long", strings.Repeat("A", 73), false, false, 4},      // 长度+小写+数字+特殊
		{"way too long", strings.Repeat("A", 100), false, false, 4}, // 长度+小写+数字+特殊

		// 单字符测试
		{"single character", "a", true, false, 4},
		{"special characters only", "!@#$%^&*()", true, false, 3},
		{"mixed content", "Test123!@#", true, true, 0},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// 测试密码哈希生成
			hash, err := GeneratePasswordHash(tt.password)
			if tt.shouldHash {
				if err != nil {
					t.Errorf("GeneratePasswordHash(%q) should not error, got: %v", tt.password, err)
				}
				if hash == "" {
					t.Errorf("GeneratePasswordHash(%q) should return non-empty hash", tt.password)
				}

				// 测试密码验证
				if flag, err := VerifyPassword(tt.password, hash); flag {
					t.Errorf("VerifyPassword(%q, hash) should return nil error, got: %v", tt.password, err)
				}

				// 验证错误密码
				if flag, err := VerifyPassword("wrongpassword", hash); flag {
					t.Errorf("VerifyPassword(wrongpassword, hash) should return error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("GeneratePasswordHash(%q) should error", tt.password)
				}
			}

			// 3. 测试密码强度检查
			strong := IsStrongPassword(tt.password)
			if strong != tt.isStrong {
				t.Errorf("IsStrongPassword(%q) = %v; want %v", tt.password, strong, tt.isStrong)
			}

			// 4. 测试详细强度验证
			valid, issues := ValidatePasswordStrength(tt.password)
			if valid != tt.isStrong {
				t.Errorf("ValidatePasswordStrength(%q) valid = %v; want %v", tt.password, valid, tt.isStrong)
			}

			if len(issues) != tt.expectedIssues {
				t.Errorf("ValidatePasswordStrength(%q) issues count = %d; want %d. Issues: %v",
					tt.password, len(issues), tt.expectedIssues, issues)
			}
		})
	}
}

// 测试密码长度边界
func TestPasswordLength(t *testing.T) {
	lengthTests := []struct {
		length            int
		shouldBeStrong    bool
		expectLengthIssue bool
	}{
		{0, false, true},
		{1, false, true},
		{7, false, true},
		{8, true, false}, // 恰好8位且符合强度要求
		{20, true, false},
		{50, true, false},
		{72, true, false}, // bcrypt最大长度
		{73, false, true}, // 超过bcrypt限制
		{100, false, true},
	}

	for _, tt := range lengthTests {
		t.Run(fmt.Sprintf("length_%d", tt.length), func(t *testing.T) {
			// 构造符合强度要求的密码
			password := ""
			if tt.length > 0 {
				base := "Aa1!"
				if tt.length >= 4 {
					password = base + strings.Repeat("a", tt.length-4)
				} else {
					password = base[:tt.length]
				}
			}

			// 测试强度
			got := IsStrongPassword(password)
			if got != tt.shouldBeStrong {
				t.Errorf("IsStrongPassword(length=%d) = %v; want %v", tt.length, got, tt.shouldBeStrong)
			}

			// 测试详细验证
			valid, issues := ValidatePasswordStrength(password)
			if valid != tt.shouldBeStrong {
				t.Errorf("ValidatePasswordStrength(length=%d) valid = %v; want %v", tt.length, valid, tt.shouldBeStrong)
			}

			// 检查是否有长度问题
			hasLengthIssue := false
			for _, issue := range issues {
				if strings.Contains(issue, "长度") {
					hasLengthIssue = true
					break
				}
			}

			if hasLengthIssue != tt.expectLengthIssue {
				t.Errorf("ValidatePasswordStrength(length=%d) length issue = %v; want %v. Issues: %v",
					tt.length, hasLengthIssue, tt.expectLengthIssue, issues)
			}
		})
	}
}

// 测试边界情况和错误处理
func TestPasswordEdgeCases(t *testing.T) {
	edgeCases := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{"correct password", "testpassword123", "", true}, // hash 会在测试中生成
		{"wrong password", "wrongpassword", "", false},
		{"empty input", "", "", false},
		{"case sensitive", "testpassword", "", false}, // 原密码是 testpassword123
		{"invalid hash", "password", "invalid_hash", false},
		{"empty hash", "password", "", false},
		{"non-bcrypt hash", "password", "plaintext_password", false},
	}

	// 先生成一个有效的哈希用于测试
	testPassword := "testpassword123"
	validHash, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			hash := tt.hash
			if hash == "" && tt.expected {
				hash = string(validHash)
			}

			if flag, err := VerifyPassword(tt.password, hash); flag {
				got := err == nil
				if got != tt.expected {
					t.Errorf("VerifyPassword(%q, %q) = %v; want %v", tt.password, hash, got, tt.expected)
				}
			}
		})
	}
}

// 性能测试
func BenchmarkPasswordFunctions(b *testing.B) {
	password := "MyTestPassword123!"
	hash, _ := GeneratePasswordHash(password)

	b.Run("VerifyPassword", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			VerifyPassword(password, hash)
		}
	})

	b.Run("GeneratePasswordHash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GeneratePasswordHash(password)
		}
	})

	b.Run("IsStrongPassword", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsStrongPassword(password)
		}
	})

	b.Run("ValidatePasswordStrength", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ValidatePasswordStrength(password)
		}
	})

	b.Run("GenerateRandomCode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateRandomCode()
		}
	})
}

// 测试随机验证码的随机性
func TestGenerateRandomCode_Randomness(t *testing.T) {
	codes := make(map[string]bool)

	// 生成100个验证码
	for i := 0; i < 100; i++ {
		code := GenerateRandomCode()
		codes[code] = true
	}

	// 至少应该有80%的验证码是不同的
	if len(codes) < 80 {
		t.Errorf("GenerateRandomCode should generate diverse codes, got %d unique codes out of 100", len(codes))
	}
}
