package model

import (
	"regexp"
	"strings"
	"time"

	"github.com/Hermitf/the-pass/pkg/formatting"
)

// 简单的验证正则
var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

// #region 模型定义

// User 用户模型
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement;comment:用户ID"`
	Username     string    `json:"username" gorm:"unique;not null;size:50;comment:用户名"`
	PasswordHash string    `json:"password_hash" gorm:"not null;size:255;comment:用户密码"`
	Email        string    `json:"email" gorm:"unique;not null;size:100;comment:用户邮箱"`
	Phone        string    `json:"phone" gorm:"unique;not null;size:11;comment:用户手机号"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	AvatarURL    string    `json:"avatar_url" gorm:"size:255;comment:用户头像URL"`
	IsActive     bool      `json:"is_active" gorm:"default:true;comment:用户是否激活"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

// #endregion

// #region 响应DTO

// UserResponse 用户响应DTO（用于API返回）
type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Name     string `json:"name"`             // 显示名称，使用username
	Avatar   string `json:"avatar,omitempty"` // 头像URL
	IsActive bool   `json:"is_active"`        // 账号状态
}

// ToResponse 将 User 模型转换为响应DTO
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Phone:    u.Phone,
		Name:     u.Username, // User 模型暂无单独的 Name 字段，使用 Username
		Avatar:   u.AvatarURL,
		IsActive: u.IsActive,
	}
}

// #endregion

// #region 验证方法

// ValidateUsername 验证用户名格式
func (u *User) ValidateUsername() error {
	if len(u.Username) < 3 {
		return ErrInvalidUsername
	}
	if len(u.Username) > 50 {
		return ErrUsernameTooLong
	}
	// 用户名只能包含字母、数字、下划线和中文（使用 Unicode 属性以避免 \u 转义）
	matched, _ := regexp.MatchString(`^[\p{L}\p{N}_]+$`, u.Username)
	if !matched {
		return ErrInvalidUsernameFormat
	}
	return nil
}

// ValidateEmail 验证邮箱格式
func (u *User) ValidateEmail() error {
	if u.Email == "" {
		return ErrEmailRequired
	}
	email := strings.TrimSpace(u.Email)
	if len(email) > 254 || !emailRegex.MatchString(email) {
		return ErrInvalidEmailFormat
	}
	return nil
}

// ValidatePhone 验证手机号格式
func (u *User) ValidatePhone() error {
	if u.Phone == "" {
		return ErrPhoneRequired
	}
	// 清理格式
	phone := strings.ReplaceAll(u.Phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	if len(phone) != 11 || !phoneRegex.MatchString(phone) {
		return ErrInvalidPhoneFormat
	}
	return nil
}

// ValidateAll 验证所有字段
func (u *User) ValidateAll() error {
	if err := u.ValidateUsername(); err != nil {
		return err
	}
	if err := u.ValidateEmail(); err != nil {
		return err
	}
	if err := u.ValidatePhone(); err != nil {
		return err
	}
	return nil
}

// #endregion

// #region 业务方法

// IsEmailTaken 检查邮箱是否已被使用（需要数据库查询，通常在service层实现）
func (u *User) IsEmailTaken() bool {
	// 这个方法的具体实现应该在 service 层
	// 这里只是为了展示方法结构
	return false
}

// IsPhoneTaken 检查手机号是否已被使用
func (u *User) IsPhoneTaken() bool {
	// 这个方法的具体实现应该在 service 层
	return false
}

// GetDisplayName 获取显示名称
func (u *User) GetDisplayName() string {
	if u.Username != "" {
		return u.Username
	}
	return u.Email
}

// GetAvatarURL 获取头像URL，如果没有则返回默认头像
func (u *User) GetAvatarURL() string {
	if u.AvatarURL != "" {
		return u.AvatarURL
	}
	// 返回默认头像URL或根据用户ID生成头像
	return "/static/avatars/default.png"
}

// UpdateProfile 更新用户基本信息
func (u *User) UpdateProfile(username, email, phone string) error {
	// 更新前验证
	tempUser := &User{
		Username: username,
		Email:    email,
		Phone:    phone,
	}

	if err := tempUser.ValidateAll(); err != nil {
		return err
	}

	u.Username = username
	u.Email = email
	u.Phone = phone
	u.UpdatedAt = time.Now()

	return nil
}

// Activate 激活用户账号
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// Deactivate 停用用户账号
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// #endregion

// #region 工具方法

// MaskEmail 脱敏显示邮箱
func (u *User) MaskEmail() string {
	return formatting.MaskEmail(u.Email)
}

// MaskPhone 脱敏显示手机号
func (u *User) MaskPhone() string {
	return formatting.MaskPhone(u.Phone)
}

// ToSafeResponse 转换为安全的响应（脱敏）
func (u *User) ToSafeResponse() *UserResponse {
	return &UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.MaskEmail(),
		Phone:    u.MaskPhone(),
		Name:     u.GetDisplayName(),
		Avatar:   u.GetAvatarURL(),
		IsActive: u.IsActive,
	}
}

// #endregion
