// Package sms 提供短信验证码的完整解决方案
//
// # 架构设计
//
// 本包采用分层架构，职责清晰分离：
//
//	┌─────────────────────────────────────────┐
//	│         Service (业务编排层)             │  ← 上层调用入口
//	│  整合限流、存储、发送的完整业务流程        │
//	└─────────────────────────────────────────┘
//	              ↓           ↓
//	    ┌─────────────┐  ┌──────────┐
//	    │   Store     │  │ Provider │          ← 抽象接口层
//	    │  (存储接口)  │  │(发送接口) │
//	    └─────────────┘  └──────────┘
//	         ↓                ↓
//	  ┌───────────┐    ┌──────────────┐       ← 实现层
//	  │RedisStore │    │ MockProvider │
//	  │           │    │ (可扩展为     │
//	  │           │    │ 阿里云/腾讯云)│
//	  └───────────┘    └──────────────┘
//
// # 核心组件
//
// 1. Store 接口 (sms.go)
//   - 定义验证码存储、读取、删除的抽象
//   - 定义频率限制和每日统计的抽象
//
// 2. RedisStore 实现 (store_redis.go)
//   - 基于 Redis 的高性能存储实现
//   - 使用 ZSET 实现滑动时间窗口限流
//   - 自动过期管理，节省内存
//
// 3. Provider 接口 (provider.go)
//   - 定义短信发送服务的抽象
//   - 支持多种第三方服务商（阿里云、腾讯云等）
//
// 4. Service 编排层 (service.go)
//   - 整合 Store 和 Provider
//   - 实现发送验证码的完整流程（限流→生成→存储→发送）
//   - 实现验证验证码的流程（读取→比对→删除）
//
// 5. 错误定义 (errors.go)
//   - 统一管理业务错误
//   - 使用哨兵错误模式，便于上层判断
//
// # 使用示例
//
// 初始化服务：
//
//	import (
//	    "github.com/Hermitf/the-pass/pkg/sms"
//	    "github.com/redis/go-redis/v9"
//	)
//
//	// 1. 创建 Redis 存储
//	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//	store := sms.NewRedisStore(redisClient)
//
//	// 2. 创建发送服务商（开发环境使用 Mock）
//	provider := sms.NewMockProvider()
//
//	// 3. 配置运行时参数
//	cfg := sms.SMSRuntimeConfig{
//	    Enabled:    true,
//	    ExpireIn:   5 * time.Minute,  // 验证码 5 分钟过期
//	    RateMax:    1,                 // 60 秒内最多发送 1 次
//	    RateWindow: 60 * time.Second,
//	    DailyMax:   10,                // 每天最多 10 次
//	    Template:   "您的验证码是 %s，5分钟内有效",
//	}
//
//	// 4. 创建服务实例
//	smsService := sms.NewService(store, provider, cfg)
//
// 发送验证码：
//
//	err := smsService.SendCode(ctx, "13800000000")
//	if err != nil {
//	    switch {
//	    case errors.Is(err, sms.ErrSendTooFrequent):
//	        // 提示用户发送过于频繁
//	    case errors.Is(err, sms.ErrDailyLimitReached):
//	        // 提示用户超过每日上限
//	    default:
//	        // 其他错误处理
//	    }
//	}
//
// 验证验证码：
//
// 检查发送可用性（不会消耗限流额度）：
//
//	 allowed, retryAfter, err := smsService.CanSend(ctx, "13800000000")
//	 if err != nil {
//	     // 处理 ErrSendTooFrequent / ErrDailyLimitReached / ErrProviderDisabled
//	 }
//	 if !allowed {
//	     // 根据 retryAfter 告诉用户需等待多久
//	 }
//
//		err := smsService.VerifyCode(ctx, "13800000000", "123456")
//		if err != nil {
//		    switch {
//		    case errors.Is(err, sms.ErrCodeExpired):
//		        // 提示验证码已过期
//		    case errors.Is(err, sms.ErrCodeMismatch):
//		        // 提示验证码错误
//		    }
//		}
//
// # 扩展指南
//
// ## 添加新的短信服务商
//
// 实现 Provider 接口即可：
//
//	type AliyunProvider struct {
//	    accessKey string
//	    secretKey string
//	}
//
//	func (a *AliyunProvider) SendSMS(ctx context.Context, phone, content string) error {
//	    // 调用阿里云 SDK 发送短信
//	    return nil
//	}
//
// ## 替换存储实现
//
// 实现 Store 接口，例如使用 MongoDB：
//
//	type MongoStore struct {
//	    collection *mongo.Collection
//	}
//
//	func (m *MongoStore) SaveCode(phone, code string, expireIn time.Duration) error {
//	    // 使用 MongoDB 存储验证码
//	    return nil
//	}
//
// # 最佳实践
//
// 1. 频率限制配置建议：
//   - RateWindow: 60 秒
//   - RateMax: 1 次
//   - DailyMax: 5-10 次
//
// 2. 验证码有效期建议：
//   - 登录/注册场景: 5 分钟
//   - 敏感操作场景: 3 分钟
//
// 3. 错误处理：
//   - 使用 errors.Is 判断具体错误类型
//   - 给用户友好的错误提示
//   - 记录详细日志便于排查
//
// 4. 安全建议：
//   - 验证成功后立即删除验证码（防重放）
//   - 使用 crypto/rand 生成验证码（防预测）
//   - 配置合理的频率限制（防刷）
package sms
