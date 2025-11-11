// Package sms 提供短信验证码相关功能
//
// 架构说明：
//   - Store 接口：定义验证码存储与限流的抽象（实现见 store_redis.go）
//   - Provider 接口：定义短信发送服务抽象（实现见 provider.go）
//   - Service：业务编排层，整合存储、发送、限流（见 service.go）
//
// 使用方式：
//
//	store := sms.NewRedisStore(redisClient)
//	provider := sms.NewMockProvider()
//	cfg := sms.SMSRuntimeConfig{...}
//	service := sms.NewService(store, provider, cfg)
//	service.SendCode(ctx, phone)
//	service.VerifyCode(ctx, phone, code)
package sms

import (
	"context"
	"crypto/rand"
	"io"
	"time"
)

// Store 定义验证码存储与限流的抽象接口
//
// 实现类需要提供：
//   - 验证码的保存、读取、删除（支持过期时间）
//   - 发送频率限制（滑动时间窗口）
//   - 每日发送次数统计
type Store interface {
	// SaveCode 保存验证码到存储，并设置过期时间
	SaveCode(phone string, code string, expireIn time.Duration) error

	// GetCode 根据手机号获取存储的验证码
	// 返回空字符串或错误表示验证码不存在或已过期
	GetCode(phone string) (string, error)

	// DeleteCode 删除手机号对应的验证码（验证成功后调用）
	DeleteCode(phone string) error

	// CheckRateLimit 检查手机号在指定时间窗口内的发送次数是否超过上限
	// maxCount: 时间窗口内最大允许次数，<=0 表示不限制
	// interval: 时间窗口大小
	// 返回 true 表示允许发送，false 表示超过限制
	CheckRateLimit(phone string, maxCount int, interval time.Duration) (bool, error)

	// IncrementDailyCount 增加手机号当天的发送计数并返回当前次数
	// 用于每日发送上限控制
	IncrementDailyCount(phone string) (int, error)

	// GetDailyCount 获取当天发送次数及剩余有效期（用于只读检测）
	GetDailyCount(phone string) (count int, ttl time.Duration, err error)

	// PeekRate 只读检查发送频率，不写入窗口。
	// 返回是否允许发送以及需要等待的时间（若不允许）。
	PeekRate(phone string, maxCount int, interval time.Duration) (allowed bool, retryAfter time.Duration, err error)
}

// CtxStore 是带上下文的存储接口，便于调用端传递超时/取消信号
// 建议新代码优先实现/使用该接口；旧接口仍保留做兼容
type CtxStore interface {
	// 带 ctx 的验证码写入/读取/删除
	SaveCodeCtx(ctx context.Context, phone string, code string, expireIn time.Duration) error
	GetCodeCtx(ctx context.Context, phone string) (string, error)
	DeleteCodeCtx(ctx context.Context, phone string) error

	// 带 ctx 的限流与统计
	CheckRateLimitCtx(ctx context.Context, phone string, maxCount int, interval time.Duration) (bool, error)
	IncrementDailyCountCtx(ctx context.Context, phone string) (int, error)
	GetDailyCountCtx(ctx context.Context, phone string) (count int, ttl time.Duration, err error)
	PeekRateCtx(ctx context.Context, phone string, maxCount int, interval time.Duration) (allowed bool, retryAfter time.Duration, err error)
}

// CodeLength 验证码长度常量
const CodeLength = 6

const (
	digitsAlphabet = "0123456789"
	fallbackMod    = 1000000
)

// GenerateCode 生成随机数字验证码（使用 crypto/rand 保证安全性）
//
// 生成 6 位纯数字验证码，适用于短信验证场景
func GenerateCode() string {
	buf := make([]byte, CodeLength)

	// 使用加密安全的随机数生成器
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		// 极端情况：降级到时间戳（不应发生）
		return fallbackCode()
	}

	for i := 0; i < CodeLength; i++ {
		buf[i] = digitsAlphabet[int(buf[i])%len(digitsAlphabet)]
	}

	return string(buf)
}

// fallbackCode 降级方案：基于时间戳生成验证码（仅在 crypto/rand 失败时使用）
func fallbackCode() string {
	return formatDigits(int(time.Now().UnixNano() % fallbackMod))
}

// formatDigits 将给定数字格式化为固定长度的数字字符串
func formatDigits(num int) string {
	buf := make([]byte, CodeLength)
	base := len(digitsAlphabet)
	for i := CodeLength - 1; i >= 0; i-- {
		buf[i] = digitsAlphabet[num%base]
		num /= base
	}
	return string(buf)
}
