package crypto

import (
	"log"
	"sync"
	"time"
)

// LimiterPolicy 限制策略
type LimiterPolicy struct {
	MaxAttempts int           // 允许的最大连续失败次数（>0 生效）
	Window      time.Duration // 计数窗口
	Lockout     time.Duration // 触发后锁定时长
}

type attemptRec struct {
	mu         sync.Mutex
	count      int
	windowFrom time.Time
	lockUntil  time.Time
}

var attemptMap sync.Map // key -> *attemptRec

// VerifyPasswordWithLimit 验证密码并对指定标识符（如用户名/IP）施加尝试次数限制
// - id 为空则不启用限制
// - 验证失败与锁定事件会记录轻量日志
// 流程：
// 1) 若启用限制：检查并累计当前 id 的尝试次数，必要时直接返回 ErrTooManyAttempts
// 2) 调用 VerifyPassword 进行校验
// 3) 若失败：记录轻量日志并返回错误（不重置计数）
// 4) 若成功：重置当前 id 的计数与锁定信息
func VerifyPasswordWithLimit(id string, hashedPassword, password string, policy LimiterPolicy) error {
	if id != "" && policy.MaxAttempts > 0 {
		if err := checkAndIncAttempts(id, policy); err != nil {
			return err
		}
	}

	err := VerifyPassword(hashedPassword, password)
	if err != nil {
		if id != "" && policy.MaxAttempts > 0 {
			// 留给上层更丰富的审计；这里仅输出一条轻日志
			log.Printf("password verify failed for id=%s", id)
		}
		return err
	}

	if id != "" && policy.MaxAttempts > 0 {
		resetAttempts(id)
	}
	return nil
}

func checkAndIncAttempts(id string, policy LimiterPolicy) error {
	// 步骤：
	// 1) 获取/初始化该 id 的计数记录
	// 2) 加锁保护记录
	// 3) 若处于锁定期，直接返回 ErrTooManyAttempts
	// 4) 若窗口已过期，重置窗口与计数
	// 5) 递增计数，若超过阈值则设置 lockUntil 并返回 ErrTooManyAttempts
	now := time.Now()
	recAny, _ := attemptMap.LoadOrStore(id, &attemptRec{windowFrom: now})
	rec := recAny.(*attemptRec)
	rec.mu.Lock()
	defer rec.mu.Unlock()

	if rec.lockUntil.After(now) {
		return ErrTooManyAttempts
	}
	// 窗口滚动
	if policy.Window > 0 && now.Sub(rec.windowFrom) > policy.Window {
		rec.windowFrom = now
		rec.count = 0
	}
	rec.count++
	if rec.count > policy.MaxAttempts {
		if policy.Lockout > 0 {
			rec.lockUntil = now.Add(policy.Lockout)
		} else if policy.Window > 0 { // 没有 Lockout 就让其到窗口结束
			rec.lockUntil = rec.windowFrom.Add(policy.Window)
		}
		log.Printf("password attempts locked id=%s until=%s", id, rec.lockUntil.Format(time.RFC3339))
		return ErrTooManyAttempts
	}
	return nil
}

func resetAttempts(id string) {
	// 成功验证后调用：将计数归零、清除锁定、刷新窗口起点
	if recAny, ok := attemptMap.Load(id); ok {
		rec := recAny.(*attemptRec)
		rec.mu.Lock()
		rec.count = 0
		rec.lockUntil = time.Time{}
		rec.windowFrom = time.Now()
		rec.mu.Unlock()
	}
}
