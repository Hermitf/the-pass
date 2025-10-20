package crypto

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"unicode"
)

const (
	// DefaultCost bcrypt 默认代价
	DefaultCost = bcrypt.DefaultCost
	// MinCost bcrypt 最小代价
	MinCost = bcrypt.MinCost
	// MaxCost bcrypt 最大代价
	MaxCost = bcrypt.MaxCost
	// BcryptPasswordMaxLength bcrypt 密码最大长度限制
	BcryptPasswordMaxLength = 72
)

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("密码不能为空")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword 验证密码是否正确
func VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" {
		return fmt.Errorf("哈希密码不能为空")
	}
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("密码验证失败")
	}

	return nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}

	if len(password) < 6 {
		return fmt.Errorf("密码至少需要6个字符")
	}

	if len(password) > BcryptPasswordMaxLength { // bcrypt 限制
		return fmt.Errorf("密码不能超过%d个字符", BcryptPasswordMaxLength)
	}

	// 检查密码复杂度
	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	// 至少要有大小写字母和数字
	if !hasLower {
		return fmt.Errorf("密码至少需要包含一个小写字母")
	}
	if !hasUpper {
		return fmt.Errorf("密码至少需要包含一个大写字母")
	}
	if !hasNumber {
		return fmt.Errorf("密码至少需要包含一个数字")
	}

	return nil
}
