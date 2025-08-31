package userutils

import (
	"errors"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// 验证密码
func VerifyPassword(password, hashedPassword string) (bool, error) {
	// TODO 注册需要测试密码强度，登陆不需要！
	// if !IsStrongPassword(password) {
	// 	return false, errors.New("password is not strong enough")
	// }
	// if flag, issues := ValidatePasswordStrength(password); !flag {
	// 	return false, errors.New("password: " + strings.Join(issues, "; "))
	// }
	// 使用bcrypt验证密码
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil, err
}

// 生成密码哈希
func GeneratePasswordHash(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", errors.New("password cannot be empty")
	}

	// bcrypt 限制密码长度为72字节
	if len(password) > 72 {
		return "", errors.New("password length exceeds 72 bytes")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// 检查密码强度
func IsStrongPassword(password string) bool {
	// 长度检查：8-72位（bcrypt限制）
	if len(password) < 8 || len(password) > 72 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// Validate password strength and return detailed info in English
func ValidatePasswordStrength(password string) (bool, []string) {
	var issues []string

	if len(password) < 8 {
		issues = append(issues, "password must be at least 8 characters long")
	}
	if len(password) > 72 {
		issues = append(issues, "password must not exceed 72 characters")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		issues = append(issues, "password must contain at least one uppercase letter")
	}
	if !hasLower {
		issues = append(issues, "password must contain at least one lowercase letter")
	}
	if !hasNumber {
		issues = append(issues, "password must contain at least one number")
	}
	if !hasSpecial {
		issues = append(issues, "password must contain at least one special character")
	}

	return len(issues) == 0, issues
}

// 生成随机字符串
func GenerateRandomCode() string {
	length := 6 // 默认长度为6
	const charset = "0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// 验证SMS验证码
func VerifySMS(phone, code string) error {
	if phone == "" || code == "" {
		return errors.New("phone number or code cannot be empty")
	}
	// 这里可以添加实际的短信验证码验证逻辑
	return nil // 假设验证通过
}

// 保存SMS验证码
func SaveSMSCode(phone, code string, duration time.Duration) error {
	if phone == "" || code == "" {
		return errors.New("phone number or code cannot be empty")
	}
	// 这里可以添加实际的SMS验证码保存逻辑
	return nil // 假设保存成功
}

func SendSMS(phone, code string) error {
	if phone == "" || code == "" {
		return errors.New("phone number or code cannot be empty")
	}
	// 这里可以添加实际的发送短信逻辑
	return nil // 假设发送成功
}
