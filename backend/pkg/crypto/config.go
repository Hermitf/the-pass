package crypto

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost bcrypt 默认代价
	DefaultCost = bcrypt.DefaultCost
	// MinCost bcrypt 最小代价
	MinCost = bcrypt.MinCost
	// MaxCost bcrypt 最大代价
	MaxCost = bcrypt.MaxCost
	// BcryptPasswordMaxLength bcrypt 密码最大长度限制（超过将被截断，故应用层要先限制长度）
	BcryptPasswordMaxLength = 72
)

var (
	bcryptCost atomic.Int32
	pepperStr  atomic.Value // stores string
)

func init() {
	// 默认 cost
	bcryptCost.Store(int32(DefaultCost))
	// 默认 pepper 为空
	pepperStr.Store("")
}

// SetBcryptCost 设置全局 bcrypt 代价（范围在 MinCost..MaxCost 之间）
// 流程：
// 1) 校验参数范围
// 2) 通过原子变量写入新值（线程安全，实时生效）
func SetBcryptCost(cost int) error {
	if cost < MinCost || cost > MaxCost {
		return fmt.Errorf("bcrypt cost 超出范围: %d (允许 %d~%d)", cost, MinCost, MaxCost)
	}
	bcryptCost.Store(int32(cost))
	return nil
}

// GetBcryptCost 读取当前 bcrypt 代价
func GetBcryptCost() int { return int(bcryptCost.Load()) }

// SetPepper 设置全局 pepper（为空表示不使用）
// 注意：pepper 为额外“服务器侧”秘密，应通过环境变量或密钥管理注入。
func SetPepper(p string) { pepperStr.Store(p) }

// GetPepper 获取全局 pepper
func GetPepper() string {
	v := pepperStr.Load()
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// LoadPasswordConfigFromEnv 从环境变量加载密码策略（可选）：
// THE_PASS_BCRYPT_COST（int）；THE_PASS_PASSWORD_PEPPER（string）
// 流程：
// 1) 读取 cost，若存在则解析为 int，并调用 SetBcryptCost 生效
// 2) 读取 pepper，若存在则直接 SetPepper
// 3) 任一解析失败将返回错误，不会中断进程（由调用方决定兜底策略）
func LoadPasswordConfigFromEnv() error {
	if v := os.Getenv("THE_PASS_BCRYPT_COST"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("THE_PASS_BCRYPT_COST 解析失败: %w", err)
		}
		if err := SetBcryptCost(n); err != nil {
			return err
		}
	}
	if p := os.Getenv("THE_PASS_PASSWORD_PEPPER"); p != "" {
		SetPepper(p)
	}
	return nil
}

// applyPepper 将应用侧的密码与服务器侧的 pepper 拼接（空 pepper 则原样返回）
// 说明：
// - 这是在 Hash 与 Verify 之前的统一入口；
// - Pepper 的存在可以降低彩虹表攻击风险，但要做好密钥管理与轮换策略；
// - 我们在 Verify 里做了“带 pepper 再不带 pepper”的回兼，便于无缝启用 pepper。
func applyPepper(password string) string {
	p := GetPepper()
	if p == "" {
		return password
	}
	return password + p
}
