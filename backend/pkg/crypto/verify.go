package crypto

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// VerifyPassword 验证密码是否正确
// 说明：
// - 支持 pepper 前后兼容：先尝试带 pepper 的密码，再尝试不带 pepper（兼容历史哈希）。
// - 失败会打印一条轻量日志（上层可替换为可插拔 Logger）。
// 流程：
// 1) 校验入参：哈希与明文非空
// 2) 对明文密码应用 pepper 并做一次 Compare
// 3) 若失败，使用原始明文再 Compare 一次（回兼旧哈希）
// 4) 若仍失败，记录日志并返回统一错误
func VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" {
		return fmt.Errorf("哈希密码不能为空")
	}
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}

	// 先尝试带 pepper 的验证
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(applyPepper(password))); err == nil {
		return nil
	}
	// 再尝试不带 pepper（兼容历史数据）
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err == nil {
		return nil
	}
	// 失败后返回统一错误
	// 仅在失败时记录一条轻量日志（上层可按需接管）
	log.Printf("password verify failed")
	return fmt.Errorf("密码验证失败")
}
