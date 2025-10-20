package sms

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Hermitf/the-pass/pkg/validator"
)

// CodeLength 验证码长度
const CodeLength = 6

// GenerateCode 生成指定长度的数字验证码
func GenerateCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := ""
	for i := 0; i < CodeLength; i++ {
		code += strconv.Itoa(r.Intn(10))
	}
	return code
}

// Send 发送短信验证码 (模拟实现)
func Send(phone, code string) error {
	// 验证手机号
	if err := validator.ValidatePhone(phone); err != nil {
		return fmt.Errorf("手机号验证失败: %w", err)
	}

	if code == "" {
		return fmt.Errorf("验证码不能为空")
	}

	// TODO: 在这里集成真实的短信服务提供商
	// 比如阿里云短信、腾讯云短信等

	// 模拟发送延时
	time.Sleep(100 * time.Millisecond)

	// 模拟发送成功
	fmt.Printf("模拟发送短信: 手机号 %s, 验证码 %s\n", phone, code)

	return nil
}

// Verify 验证手机验证码 (基础实现)
func Verify(inputCode, expectedCode string) error {
	if inputCode == "" {
		return fmt.Errorf("请输入验证码")
	}

	if expectedCode == "" {
		return fmt.Errorf("验证码已过期")
	}

	if inputCode != expectedCode {
		return fmt.Errorf("验证码错误")
	}

	return nil
}

// IsValidCode 检查验证码格式是否正确
func IsValidCode(code string) bool {
	if len(code) != CodeLength {
		return false
	}

	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
