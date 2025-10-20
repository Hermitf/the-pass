package validator

import (
	"fmt"
	"regexp"
	"strings"
)

const maxEmailLength = 254 // RFC 5321 limit for email addresses

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

// IsEmail 验证邮箱格式
func IsEmail(email string) bool {
	if email == "" {
		return false
	}
	email = strings.TrimSpace(email)
	return len(email) <= maxEmailLength && emailRegex.MatchString(email)
}

// ValidateEmail 验证邮箱并返回错误
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("邮箱地址不能为空")
	}

	email = strings.TrimSpace(email)
	if len(email) > maxEmailLength {
		return fmt.Errorf("邮箱地址过长")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("邮箱格式不正确")
	}

	return nil
}

// IsPhone 验证手机号格式
func IsPhone(phone string) bool {
	if phone == "" {
		return false
	}
	// 清理格式
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	return phoneRegex.MatchString(phone)
}

// ValidatePhone 验证手机号并返回错误
func ValidatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("手机号不能为空")
	}

	// 清理格式
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")

	if len(cleanPhone) != 11 {
		return fmt.Errorf("手机号长度不正确")
	}

	if !phoneRegex.MatchString(cleanPhone) {
		return fmt.Errorf("手机号格式不正确")
	}

	return nil
}
