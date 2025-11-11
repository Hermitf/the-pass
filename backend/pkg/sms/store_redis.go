package sms

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore Redis 实现的短信验证码存储
//
// 功能：
//   - 验证码存储（带过期时间）
//   - 滑动时间窗口限流（使用 Redis ZSET）
//   - 每日发送次数统计
//
// 已改进：
// - 错误处理：为 Redis 操作添加了更丰富的上下文信息（操作/键名）。
// - 键命名：引入可配置前缀（env/version），便于多环境/演进；默认前缀为 "sms"。
// - 高并发优化：频率限制与每日计数改用 Lua 原子脚本，减少往返并避免竞态。
//
// Redis 键命名规范：
//   - sms:code:{phone}      验证码存储
//   - sms:rate_z:{phone}    限流时间窗口（ZSET，分值为时间戳）
//   - sms:daily:{date}:{phone}  每日计数
type RedisStore struct {
	client *redis.Client
	prefix string // 键名前缀，默认 "sms"，支持多环境如 "dev:sms" / "prod:sms"
	logger Logger // 可选日志接口，为 nil 时回退 log.Printf
}

// NewRedisStore 创建 Redis 存储实例
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client, prefix: "sms", logger: StdLogger{}}
}

// NewRedisStoreWithPrefix 创建带前缀的 Redis 存储实例（前缀末尾无需冒号）
func NewRedisStoreWithPrefix(client *redis.Client, prefix string) *RedisStore {
	if prefix == "" {
		prefix = "sms"
	}
	return &RedisStore{client: client, prefix: prefix, logger: StdLogger{}}
}

// SetLogger 设置自定义日志器
func (r *RedisStore) SetLogger(l Logger) {
	if l == nil {
		return
	}
	r.logger = l
}

// Redis 键生成函数
func (r *RedisStore) codeKey(phone string) string {
	return fmt.Sprintf("%s:code:%s", r.prefix, phone)
}

func (r *RedisStore) rateSortedSet(phone string) string {
	return fmt.Sprintf("%s:rate_z:%s", r.prefix, phone)
}

func (r *RedisStore) dailyKey(phone string) string {
	return fmt.Sprintf("%s:daily:%s:%s", r.prefix, time.Now().Format("20060102"), phone)
}

// SaveCode 保存验证码并设置过期时间（包含简易脱敏日志）
// SaveCodeCtx 使用调用方上下文
func (r *RedisStore) SaveCodeCtx(ctx context.Context, phone string, code string, expireIn time.Duration) error {
	key := r.codeKey(phone)
	if err := r.client.Set(ctx, key, code, expireIn).Err(); err != nil {
		return wrapRedisErr("SET", key, err)
	}
	// 简易日志：生产可替换为结构化日志并脱敏
	if r.logger != nil {
		r.logger.Infof("sms.SaveCode phone=%s ttl=%s", maskPhone(phone), expireIn.String())
	} else {
		log.Printf("sms.SaveCode phone=%s ttl=%s", maskPhone(phone), expireIn.String())
	}
	return nil
}

// SaveCode 兼容旧接口，使用 context.Background
func (r *RedisStore) SaveCode(phone string, code string, expireIn time.Duration) error {
	return r.SaveCodeCtx(context.Background(), phone, code, expireIn)
}

// GetCode 获取存储的验证码
//
// 如果验证码不存在或已过期，返回 redis.Nil 错误
func (r *RedisStore) GetCodeCtx(ctx context.Context, phone string) (string, error) {
	key := r.codeKey(phone)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", wrapRedisErr("GET", key, err)
	}
	return val, nil
}

func (r *RedisStore) GetCode(phone string) (string, error) {
	return r.GetCodeCtx(context.Background(), phone)
}

// DeleteCode 删除验证码（验证成功后调用）
func (r *RedisStore) DeleteCodeCtx(ctx context.Context, phone string) error {
	key := r.codeKey(phone)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return wrapRedisErr("DEL", key, err)
	}
	return nil
}

func (r *RedisStore) DeleteCode(phone string) error {
	return r.DeleteCodeCtx(context.Background(), phone)
}

// CheckRateLimit 检查发送频率限制（滑动时间窗口算法）
//
// 实现原理：
//  1. 使用 Redis ZSET 存储每次发送的时间戳（分值为纳秒时间戳）
//  2. 删除窗口外的旧记录
//  3. 添加当前时间戳
//  4. 统计窗口内的发送次数
//  5. 判断是否超过限制
//
// 参数：
//   - phone: 手机号
//   - maxCount: 时间窗口内最大允许次数（<=0 表示不限制）
//   - interval: 时间窗口大小（如 1分钟）
//
// 返回：
//   - true: 允许发送
//   - false: 超过频率限制
//
// 使用 Lua 脚本在 Redis 端原子执行（删旧+写新+计数+过期）
func (r *RedisStore) CheckRateLimitCtx(ctx context.Context, phone string, maxCount int, interval time.Duration) (bool, error) {
	if maxCount <= 0 { // 不限制
		return true, nil
	}
	zkey := r.rateSortedSet(phone)
	now := time.Now()
	nowScore := strconv.FormatFloat(float64(now.UnixNano()), 'f', -1, 64)
	windowStart := now.Add(-interval)
	windowScore := strconv.FormatFloat(float64(windowStart.UnixNano()), 'f', -1, 64)
	expireSeconds := strconv.FormatInt(int64(interval.Seconds()), 10)

	res, err := luaRateLimitScript.Run(ctx, r.client, []string{zkey}, nowScore, windowScore, strconv.Itoa(maxCount), expireSeconds).Result()
	if err != nil {
		return false, wrapRedisErr("EVAL rate_limit", zkey, err)
	}
	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return false, fmt.Errorf("redis EVAL rate_limit invalid result: %T", res)
	}
	allowed, _ := arr[0].(int64)
	// count := arr[1]
	return allowed == 1, nil
}

func (r *RedisStore) CheckRateLimit(phone string, maxCount int, interval time.Duration) (bool, error) {
	return r.CheckRateLimitCtx(context.Background(), phone, maxCount, interval)
}

// PeekRate 只读检查当前窗口内的发送次数，不写入新记录
// 返回：
//
//	allowed: 是否仍可发送（count < maxCount）
//	retryAfter: 若不允许，距离窗口最早记录过期的时间（用于前端展示等待秒数）
func (r *RedisStore) PeekRateCtx(ctx context.Context, phone string, maxCount int, interval time.Duration) (bool, time.Duration, error) {
	if maxCount <= 0 {
		return true, 0, nil
	}
	zkey := r.rateSortedSet(phone)
	now := time.Now()
	nowNs := float64(now.UnixNano())
	windowStart := now.Add(-interval)
	windowScore := strconv.FormatFloat(float64(windowStart.UnixNano()), 'f', -1, 64)
	nowScoreStr := strconv.FormatFloat(nowNs, 'f', -1, 64)

	res, err := luaPeekRateScript.Run(ctx, r.client, []string{zkey}, windowScore, strconv.Itoa(maxCount), nowScoreStr).Result()
	if err != nil {
		return false, 0, wrapRedisErr("EVAL peek_rate", zkey, err)
	}
	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return false, 0, fmt.Errorf("redis EVAL peek_rate invalid result: %T", res)
	}
	allowed, _ := arr[0].(int64)
	if allowed == 1 {
		return true, 0, nil
	}
	// 计算剩余等待时间 = earliest + interval - now
	var earliestNs float64
	switch v := arr[1].(type) {
	case int64:
		earliestNs = float64(v)
	case float64:
		earliestNs = v
	case string:
		if f, convErr := strconv.ParseFloat(v, 64); convErr == nil {
			earliestNs = f
		}
	}
	deltaNs := (earliestNs + float64(interval.Nanoseconds())) - nowNs
	if deltaNs < 0 {
		deltaNs = 0
	}
	return false, time.Duration(deltaNs) * time.Nanosecond, nil
}

func (r *RedisStore) PeekRate(phone string, maxCount int, interval time.Duration) (bool, time.Duration, error) {
	return r.PeekRateCtx(context.Background(), phone, maxCount, interval)
}

// IncrementDailyCount 增加当日发送次数并返回当前计数
//
// 用于每日发送上限控制
// 键会在每天 00:00 自动过期
func (r *RedisStore) IncrementDailyCountCtx(ctx context.Context, phone string) (int, error) {
	dkey := r.dailyKey(phone)
	expireSec := secondsUntilEndOfDay()
	if expireSec <= 0 {
		expireSec = 24 * 60 * 60
	}
	res, err := luaDailyIncrScript.Run(ctx, r.client, []string{dkey}, strconv.FormatInt(expireSec, 10)).Result()
	if err != nil {
		return 0, wrapRedisErr("EVAL daily_incr", dkey, err)
	}
	switch v := res.(type) {
	case int64:
		return int(v), nil
	case string:
		if n, convErr := strconv.Atoi(v); convErr == nil {
			return n, nil
		}
	}
	return 0, fmt.Errorf("redis EVAL daily_incr invalid result: %T", res)
}

// GetDailyCount 返回当天已发送次数及 key 剩余 TTL（若 key 不存在返回 0,0）
func (r *RedisStore) GetDailyCountCtx(ctx context.Context, phone string) (int, time.Duration, error) {
	dkey := r.dailyKey(phone)

	val, err := r.client.Get(ctx, dkey).Result()
	if err != nil {
		if err == redis.Nil {
			ttl, _ := r.client.TTL(ctx, dkey).Result()
			return 0, ttl, nil
		}
		return 0, 0, wrapRedisErr("GET", dkey, err)
	}

	count, convErr := strconv.Atoi(val)
	if convErr != nil {
		return 0, 0, convErr
	}

	ttl, err := r.client.TTL(ctx, dkey).Result()
	if err != nil && err != redis.Nil {
		return count, 0, wrapRedisErr("TTL", dkey, err)
	}
	return count, ttl, nil
}

func (r *RedisStore) IncrementDailyCount(phone string) (int, error) {
	return r.IncrementDailyCountCtx(context.Background(), phone)
}

func (r *RedisStore) GetDailyCount(phone string) (int, time.Duration, error) {
	return r.GetDailyCountCtx(context.Background(), phone)
}
