package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/Hermitf/the-pass/pkg/validator"
)

// Service 短信验证码业务服务（核心编排层）
//
// 职责：
//   - 整合存储（Store）、发送（Provider）、配置（Config）
//   - 实现发送验证码的完整流程：限流 → 生成码 → 存储 → 发送
//   - 实现验证验证码的流程：读取 → 比对 → 删除
//
// 使用示例：
//
//	store := sms.NewRedisStore(redisClient)
//	provider := sms.NewMockProvider()
//	cfg := sms.SMSRuntimeConfig{Enabled: true, ExpireIn: 5*time.Minute, ...}
//	service := sms.NewService(store, provider, cfg)
//	err := service.SendCode(ctx, "13800000000")
type Service struct {
	store    Store
	provider Provider
	cfg      SMSRuntimeConfig
}

// ensureEnabled 返回服务是否启用的错误信息
func (s *Service) ensureEnabled() error {
	if !s.cfg.Enabled {
		return ErrProviderDisabled
	}
	return nil
}

// validatePhone 检查手机号是否合法
func (s *Service) validatePhone(phone string) error {
	if phone == "" || !validator.IsPhone(phone) {
		return ErrPhoneInvalid
	}
	return nil
}

// enforceRateLimit 写入模式的限流检测
func (s *Service) enforceRateLimit(phone string) error {
	allowed, err := s.store.CheckRateLimit(phone, s.cfg.RateMax, s.cfg.RateWindow)
	if err != nil {
		return fmt.Errorf("限流检查失败: %w", err)
	}
	if !allowed {
		return ErrSendTooFrequent
	}
	return nil
}

// enforceDailyLimit 递增日发送次数并判断是否超过上限
func (s *Service) enforceDailyLimit(phone string) error {
	if s.cfg.DailyMax <= 0 {
		return nil
	}
	count, err := s.store.IncrementDailyCount(phone)
	if err != nil {
		return fmt.Errorf("每日计数失败: %w", err)
	}
	if count > s.cfg.DailyMax {
		return ErrDailyLimitReached
	}
	return nil
}

// peekRateLimit 只读模式的限流检测
func (s *Service) peekRateLimit(ctx context.Context, phone string) (bool, time.Duration, error) {
	if cs, ok := s.store.(CtxStore); ok {
		allowed, retryAfter, err := cs.PeekRateCtx(ctx, phone, s.cfg.RateMax, s.cfg.RateWindow)
		if err != nil {
			return false, 0, fmt.Errorf("限流只读检查失败: %w", err)
		}
		if !allowed {
			return false, retryAfter, ErrSendTooFrequent
		}
		return true, 0, nil
	}
	allowed, retryAfter, err := s.store.PeekRate(phone, s.cfg.RateMax, s.cfg.RateWindow)
	if err != nil {
		return false, 0, fmt.Errorf("限流只读检查失败: %w", err)
	}
	if !allowed {
		return false, retryAfter, ErrSendTooFrequent
	}
	return true, 0, nil
}

// inspectDailyLimit 读取日发送次数状态
func (s *Service) inspectDailyLimit(ctx context.Context, phone string) (bool, time.Duration, error) {
	if s.cfg.DailyMax <= 0 {
		return true, 0, nil
	}
	if cs, ok := s.store.(CtxStore); ok {
		count, ttl, err := cs.GetDailyCountCtx(ctx, phone)
		if err != nil {
			return false, 0, fmt.Errorf("每日计数查询失败: %w", err)
		}
		if count >= s.cfg.DailyMax {
			retry := ttl
			if retry < 0 {
				retry = 0
			}
			return false, retry, ErrDailyLimitReached
		}
		return true, 0, nil
	}
	count, ttl, err := s.store.GetDailyCount(phone)
	if err != nil {
		return false, 0, fmt.Errorf("每日计数查询失败: %w", err)
	}
	if count >= s.cfg.DailyMax {
		retry := ttl
		if retry < 0 {
			retry = 0
		}
		return false, retry, ErrDailyLimitReached
	}
	return true, 0, nil
}

// SMSRuntimeConfig 运行时配置（与配置文件结构解耦）
//
// 字段说明：
//   - Enabled: 是否启用短信功能（false 时所有操作返回 ErrProviderDisabled）
//   - ExpireIn: 验证码有效期（如 5 分钟）
//   - RateMax: 时间窗口内最大发送次数（如 1 次）
//   - RateWindow: 时间窗口大小（如 60 秒）
//   - DailyMax: 每日最大发送次数（0 表示不限制）
//   - Template: 短信内容模板（如 "您的验证码是 %s，5分钟内有效"）
type SMSRuntimeConfig struct {
	Enabled    bool
	ExpireIn   time.Duration
	RateMax    int
	RateWindow time.Duration
	DailyMax   int
	Template   string
}

// NewService 创建短信服务实例
func NewService(store Store, provider Provider, cfg SMSRuntimeConfig) *Service {
	return &Service{store: store, provider: provider, cfg: cfg}
}

// SendCode 发送验证码（完整流程）
//
// 执行步骤：
//  1. 检查服务是否启用
//  2. 验证手机号格式
//  3. 检查发送频率限制（防刷）
//  4. 检查每日发送上限（可选）
//  5. 生成随机验证码
//  6. 保存验证码到存储（带过期时间）
//  7. 调用 Provider 发送短信
//  8. 如果发送失败，删除已保存的验证码
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - phone: 目标手机号
//
// 返回：
//   - nil: 发送成功
//   - ErrProviderDisabled: 短信服务未启用
//   - ErrPhoneInvalid: 手机号格式错误
//   - ErrSendTooFrequent: 发送过于频繁
//   - ErrDailyLimitReached: 超过每日上限
//   - 其他错误: 存储或发送失败
func (s *Service) SendCode(ctx context.Context, phone string) error {
	// 1. 检查服务状态
	if err := s.ensureEnabled(); err != nil {
		return err
	}

	// 2. 验证手机号
	if err := s.validatePhone(phone); err != nil {
		return err
	}

	// 3. 频率限制检查
	if err := s.enforceRateLimit(phone); err != nil {
		return err
	}

	// 4. 每日上限检查（可选）
	if err := s.enforceDailyLimit(phone); err != nil {
		return err
	}

	// 5. 生成验证码
	code := GenerateCode()

	// 6. 保存到存储（优先使用带 ctx 的接口）
	if cs, ok := s.store.(CtxStore); ok {
		if err := cs.SaveCodeCtx(ctx, phone, code, s.cfg.ExpireIn); err != nil {
			return fmt.Errorf("验证码保存失败: %w", err)
		}
	} else {
		if err := s.store.SaveCode(phone, code, s.cfg.ExpireIn); err != nil {
			return fmt.Errorf("验证码保存失败: %w", err)
		}
	}

	// 7. 发送短信
	content := FormatContent(s.cfg.Template, code)
	if err := s.provider.SendSMS(ctx, phone, content); err != nil {
		// 发送失败则删除已保存的验证码（忽略删除错误）
		if cs, ok := s.store.(CtxStore); ok {
			_ = cs.DeleteCodeCtx(ctx, phone)
		} else {
			_ = s.store.DeleteCode(phone)
		}
		return fmt.Errorf("短信发送失败: %w", err)
	}

	return nil
}

// VerifyCode 验证验证码
//
// 执行步骤：
//  1. 检查验证码是否为空
//  2. 从存储中获取保存的验证码
//  3. 比对验证码
//  4. 验证成功后删除验证码（防止重复使用）
//
// 参数：
//   - ctx: 上下文
//   - phone: 手机号
//   - code: 用户输入的验证码
//
// 返回：
//   - nil: 验证成功
//   - ErrCodeEmpty: 验证码为空
//   - ErrCodeExpired: 验证码不存在或已过期
//   - ErrCodeMismatch: 验证码错误
func (s *Service) VerifyCode(ctx context.Context, phone, code string) error {
	// 1. 参数检查
	if code == "" {
		return ErrCodeEmpty
	}

	// 2. 获取存储的验证码（优先使用带 ctx 的接口）
	var stored string
	var err error
	if cs, ok := s.store.(CtxStore); ok {
		stored, err = cs.GetCodeCtx(ctx, phone)
	} else {
		stored, err = s.store.GetCode(phone)
	}
	if err != nil || stored == "" {
		return ErrCodeExpired
	}

	// 3. 比对验证码
	if stored != code {
		return ErrCodeMismatch
	}

	// 4. 验证成功，删除验证码（一次性使用）
	if cs, ok := s.store.(CtxStore); ok {
		_ = cs.DeleteCodeCtx(ctx, phone)
	} else {
		_ = s.store.DeleteCode(phone)
	}
	return nil
}

// CanSend 只读检测：当前是否允许发送验证码，并返回需要等待的时间
// 不会写入限流窗口，适合前端“按钮冷却时间”展示
func (s *Service) CanSend(ctx context.Context, phone string) (bool, time.Duration, error) {
	if err := s.ensureEnabled(); err != nil {
		return false, 0, err
	}
	if err := s.validatePhone(phone); err != nil {
		return false, 0, err
	}

	if allowed, retryAfter, err := s.peekRateLimit(ctx, phone); err != nil || !allowed {
		return false, retryAfter, err
	}

	if allowed, retryAfter, err := s.inspectDailyLimit(ctx, phone); err != nil || !allowed {
		return false, retryAfter, err
	}

	return true, 0, nil
}
