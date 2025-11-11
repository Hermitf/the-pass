package sms

import (
	"context"
	"fmt"
	"log"
)

// Provider 短信发送服务抽象接口
//
// 用于对接不同的第三方短信服务商（阿里云、腾讯云等）
// 实现类需要处理实际的 HTTP 请求、认证、错误重试等逻辑
type Provider interface {
	// SendSMS 发送短信到指定手机号
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - phone: 目标手机号（已验证格式）
	//   - content: 短信内容（已渲染完成的文本）
	//
	// 返回：
	//   - error: 发送失败时返回错误（网络、限流、余额不足等）
	SendSMS(ctx context.Context, phone string, content string) error
}

// MockProvider 模拟短信发送实现（用于开发与测试阶段）
//
// 不会真正发送短信，只打印日志到控制台
// 生产环境请替换为真实的 Provider 实现
type MockProvider struct{}

// NewMockProvider 创建 Mock Provider 实例
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// SendSMS 模拟发送，仅打印日志
func (m *MockProvider) SendSMS(ctx context.Context, phone string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		log.Printf("[SMS Mock] 发送到 %s，内容：%s", phone, content)
		return nil
	}
}

// FormatContent 根据模板渲染短信内容
//
// 参数：
//   - templateCode: 短信模板（如 "您的验证码是 %s，5分钟内有效"）
//   - code: 验证码
//
// 返回：渲染后的短信文本
//
// 如果模板为空，使用默认格式
func FormatContent(templateCode, code string) string {
	if templateCode == "" {
		return fmt.Sprintf("您的验证码是 %s，请在有效期内使用。", code)
	}
	return fmt.Sprintf(templateCode, code)
}
