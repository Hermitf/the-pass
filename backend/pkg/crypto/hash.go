package crypto

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 加密密码
// 支持动态调整 cost（通过 SetBcryptCost/LoadPasswordConfigFromEnv），并支持 pepper（可通过 SetPepper 配置）。
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("密码不能为空")
	}

	// 读取当前生效的 cost 与 pepper
	cost := GetBcryptCost()
	pw := applyPepper(password)

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pw), cost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}

	return string(hashedBytes), nil
}

// ValidatePassword 验证密码强度
// 规则：
// - 长度：6 ~ 72（bcrypt 上限，避免被截断）
// - 复杂度：至少包含大小写字母与数字
// 使用时机：注册/修改密码前置校验，尽量在进入 Hash 前就拦截弱密码。
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}

	if len(password) < 6 {
		return fmt.Errorf("密码至少需要6个字符")
	}

	if len(password) > BcryptPasswordMaxLength { // bcrypt 限制
		return fmt.Errorf("密码不能超过%d个字符", BcryptPasswordMaxLength)
	}

	// 检查密码复杂度
	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	// 至少要有大小写字母和数字
	if !hasLower {
		return fmt.Errorf("密码至少需要包含一个小写字母")
	}
	if !hasUpper {
		return fmt.Errorf("密码至少需要包含一个大写字母")
	}
	if !hasNumber {
		return fmt.Errorf("密码至少需要包含一个数字")
	}

	return nil
}
