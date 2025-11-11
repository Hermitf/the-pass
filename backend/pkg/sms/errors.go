package sms

import "errors"

// 短信业务错误定义（集中管理）
//
// 使用 errors.New 创建哨兵错误，便于上层使用 errors.Is 判断
//
// 使用示例：
//
//	if errors.Is(err, sms.ErrSendTooFrequent) {
//	    // 处理频率限制错误
//	}
var (
	// ErrPhoneInvalid 手机号格式不正确
	ErrPhoneInvalid = errors.New("手机号格式不正确")

	// ErrCodeEmpty 验证码为空
	ErrCodeEmpty = errors.New("验证码为空")

	// ErrCodeExpired 验证码已过期或不存在
	ErrCodeExpired = errors.New("验证码已过期或不存在")

	// ErrCodeMismatch 验证码不匹配
	ErrCodeMismatch = errors.New("验证码不匹配")

	// ErrSendTooFrequent 发送过于频繁（触发频率限制）
	ErrSendTooFrequent = errors.New("发送过于频繁，请稍后再试")

	// ErrDailyLimitReached 当天发送次数已达上限
	ErrDailyLimitReached = errors.New("当天发送次数已达上限")

	// ErrProviderDisabled 短信服务未启用
	ErrProviderDisabled = errors.New("短信服务未启用")

	// ErrStoreFailure 短信存储访问失败（统一包装 Redis 之类的后端错误）
	// 上层可用 errors.Is(err, ErrStoreFailure) 判断是否为存储层异常
	ErrStoreFailure = errors.New("短信存储访问失败")
)
