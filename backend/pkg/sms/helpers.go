package sms

import (
	"errors"
	"fmt"
	"math"
	"time"
)

func maskPhone(phone string) string {
	// 简易脱敏：保留前三后四
	n := len(phone)
	if n <= 4 {
		return phone
	}
	head := 3
	if n < 7 { // 太短时减少头保留位
		head = int(math.Max(0, float64(n-4)))
	}
	return phone[:head] + "****" + phone[n-4:]
}

func wrapRedisErr(op, key string, err error) error {
	if err == nil {
		return nil
	}
	// 包装底层错误并附加哨兵 ErrStoreFailure
	return errors.Join(ErrStoreFailure, fmt.Errorf("redis %s key=%s: %w", op, key, err))
}

// 计算距离当天 23:59:59 的剩余秒数
func secondsUntilEndOfDay() int64 {
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	if !end.After(now) {
		end = end.Add(24 * time.Hour)
	}
	return int64(end.Sub(now).Seconds())
}
