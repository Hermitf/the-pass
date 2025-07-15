package userutils

import (
	"regexp"
	"strings"
)

// 预编译正则表达式，提高性能
var (
	phoneRegex   = regexp.MustCompile(`^1[3-9]\d{9}$`)
	mobileRegex  = regexp.MustCompile(`^1(3[4-9]|47|5[0-2]|5[7-9]|72|78|8[2-4]|8[7-8]|98)\d{8}$`)
	unicomRegex  = regexp.MustCompile(`^1(3[0-2]|45|5[5-6]|66|71|7[5-6]|8[5-6])\d{8}$`)
	telecomRegex = regexp.MustCompile(`^1(33|49|53|73|77|8[0-1]|89|91|99)\d{8}$`)
)

// IsPhone 验证手机号格式
func IsPhone(input string) bool {
	// 去除前后空格
	phone := strings.TrimSpace(input)

	// 使用正则表达式验证
	return phoneRegex.MatchString(phone)
}

// 检查是否为移动号码
func IsMobilePhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	return mobileRegex.MatchString(phone)
}

// 检查是否为联通号码
func IsUnicomPhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	return unicomRegex.MatchString(phone)
}

// 检查是否为电信号码
func IsTelecomPhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	return telecomRegex.MatchString(phone)
}
