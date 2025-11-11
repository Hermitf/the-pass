package sms

import (
	"log"
)

// Logger 抽象日志接口（简化版）
// 可通过适配 zap/logrus 等日志库实现该接口
// 建议仅用于关键业务点的结构化输出

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

// StdLogger 使用标准库 log 作为默认实现
// 注意：仅作为占位，生产建议替换为结构化日志

type StdLogger struct{}

func (StdLogger) Infof(format string, args ...any) {
	log.Printf(format, args...)
}

func (StdLogger) Errorf(format string, args ...any) {
	log.Printf(format, args...)
}
