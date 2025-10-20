package service

import (
	"fmt"
	"log"
	"time"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/repository"
	"github.com/Hermitf/the-pass/pkg/crypto"
	"github.com/Hermitf/the-pass/pkg/sms"
	"github.com/Hermitf/the-pass/pkg/validator"
)

// #region 服务定义

// UserServiceInterface 用户服务接口
type UserServiceInterface interface {
	// 用户注册和认证
	RegisterUser(user *model.User) error
	LoginUser(loginInfo, password, loginType string) (string, error)

	// 短信验证相关
	SendSMSCode(phone string) error
	VerifySMSCode(phone, code string) error

	// 用户信息管理
	GetUserProfile(userID uint) (*model.User, error)
	GetUserByID(userID int64) (*model.User, error)
	UpdateUserProfile(userID uint, username, email, phone string) error
	UpdatePassword(userID uint, oldPassword, newPassword string) error
	ResetPassword(identifier, newPassword string) error

	// 用户验证
	ValidateUserData(user *model.User) error
	CheckUserAvailability(username, email, phone string) error

	// 用户列表和搜索
	GetUserList(offset, limit int) ([]*model.User, int64, error)
	SearchUsers(keyword string, offset, limit int) ([]*model.User, int64, error)

	// 用户统计
	GetUserStats() (map[string]interface{}, error)
}

// UserService 用户服务实现
type UserService struct {
	userRepo   repository.UserRepositoryInterface
	jwtService JWTServiceInterface
}

// #endregion

// #region 构造函数和依赖注入

// UserServiceDependencies 用户服务依赖
type UserServiceDependencies struct {
	UserRepo   repository.UserRepositoryInterface
	JWTService JWTServiceInterface
}

// NewUserService 创建用户服务实例
func NewUserService(deps UserServiceDependencies) UserServiceInterface {
	return &UserService{
		userRepo:   deps.UserRepo,
		jwtService: deps.JWTService,
	}
}

// #endregion

// #region 用户注册和认证

// RegisterUser 注册用户
func (s *UserService) RegisterUser(user *model.User) error {
	if user == nil {
		return ErrUserNil
	}

	// 验证用户数据
	if err := s.ValidateUserData(user); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 检查用户是否已存在
	if err := s.CheckUserAvailability(user.Username, user.Email, user.Phone); err != nil {
		return fmt.Errorf("%w: %v", ErrAvailabilityCheck, err)
	}

	// 加密密码
	if user.PasswordHash != "" {
		hashedPassword, err := crypto.HashPassword(user.PasswordHash)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPasswordHashing, err)
		}
		user.PasswordHash = hashedPassword
	}

	// 创建用户
	if err := s.userRepo.CreateUser(user); err != nil {
		return fmt.Errorf("%w: %v", ErrUserCreationFailed, err)
	}

	s.logUserRegistered(user)
	return nil
}

// LoginUser 用户登录
func (s *UserService) LoginUser(loginInfo, password, loginType string) (string, error) {
	if loginInfo == "" || password == "" {
		return "", ErrLoginInfoEmpty
	}

	// 设置默认登录类型
	if loginType == "" {
		loginType = "password"
	}

	// 根据登录信息类型获取用户
	user, err := s.getUserByLoginInfo(loginInfo)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}

	// 验证登录凭据
	if err := s.verifyLoginCredentials(user, loginInfo, password, loginType); err != nil {
		return "", ErrInvalidCredentials
	}

	// 生成 JWT Token
	token, err := s.generateToken(user)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	s.logUserLogin(user, loginType)
	return token, nil
}

// #endregion

// #region 短信验证相关

// SendSMSCode 发送短信验证码
func (s *UserService) SendSMSCode(phone string) error {
	if phone == "" {
		return ErrPhoneEmpty
	}

	// 验证手机号格式
	if !validator.IsPhone(phone) {
		return ErrPhoneInvalid
	}

	// 检查用户是否存在
	user, err := s.userRepo.GetUserByPhone(phone)
	if err != nil {
		return ErrPhoneNotRegistered
	}

	// 检查发送频率限制
	if err := s.checkSMSRateLimit(phone); err != nil {
		return err
	}

	// 生成并发送短信验证码
	code := sms.GenerateCode()
	if err := sms.Send(phone, code); err != nil {
		return ErrSMSSendFailed
	}

	// 记录发送日志
	s.logSMSSent(phone, user.ID)
	return nil
}

// VerifySMSCode 验证短信验证码
func (s *UserService) VerifySMSCode(phone, code string) error {
	if phone == "" || code == "" {
		return ErrSMSCodeEmpty
	}

	// 从缓存或数据库获取实际的验证码进行比较
	actualCode, err := s.getStoredSMSCode(phone)
	if err != nil {
		return err
	}
	if err := sms.Verify(code, actualCode); err != nil {
		return ErrSMSCodeInvalid
	}
	return nil
}

// #endregion

// #region 用户信息管理

// GetUserProfile 获取用户档案
func (s *UserService) GetUserProfile(userID uint) (*model.User, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetUserByID 根据ID获取用户信息（int64版本）
func (s *UserService) GetUserByID(userID int64) (*model.User, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	return s.GetUserProfile(uint(userID))
}

// UpdateUserProfile 更新用户档案
func (s *UserService) UpdateUserProfile(userID uint, username, email, phone string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}

	// 获取当前用户信息
	currentUser, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}

	// 准备更新数据
	updateData := &model.User{
		ID:       currentUser.ID,
		Username: username,
		Email:    email,
		Phone:    phone,
		// 保留原有数据
		PasswordHash: currentUser.PasswordHash,
		CreatedAt:    currentUser.CreatedAt,
	}

	// 业务验证：验证新的用户信息格式
	if err := s.validateUserFields(updateData); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 业务验证：检查重复数据（排除当前用户）
	if err := s.checkUserAvailabilityExcluding(userID, username, email, phone); err != nil {
		return fmt.Errorf("%w: %v", ErrAvailabilityCheck, err)
	}

	// 执行更新（使用基础的Repository方法）
	if err := s.userRepo.UpdateUser(updateData); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	// 业务日志记录
	s.logUserProfileUpdated(userID)
	return nil
}

// UpdatePassword 更新用户密码
func (s *UserService) UpdatePassword(userID uint, oldPassword, newPassword string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	if oldPassword == "" || newPassword == "" {
		return ErrPasswordsEmpty
	}

	// 获取用户信息
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// 验证旧密码
	if err := crypto.VerifyPassword(user.PasswordHash, oldPassword); err != nil {
		return ErrOldPasswordIncorrect
	}

	// 加密新密码
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return ErrPasswordHashing
	}

	// 更新密码
	user.PasswordHash = hashedPassword
	if err := s.userRepo.UpdateUser(user); err != nil {
		return ErrUserUpdateFailed
	}

	s.logPasswordUpdated(userID)
	return nil
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(identifier, newPassword string) error {
	if identifier == "" || newPassword == "" {
		return ErrLoginInfoEmpty
	}

	// 根据标识符获取用户
	user, err := s.getUserByLoginInfo(identifier)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}

	// 加密新密码
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPasswordHashing, err)
	}

	// 更新密码
	user.PasswordHash = hashedPassword
	if err := s.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logPasswordReset(user.ID)
	return nil
}

// #endregion

// #region 用户验证

// ValidateUserData 验证用户数据
func (s *UserService) ValidateUserData(user *model.User) error {
	if user == nil {
		return ErrUserNil
	}

	return s.validateUserFields(user)
}

// CheckUserAvailability 检查用户可用性
func (s *UserService) CheckUserAvailability(username, email, phone string) error {
	// 业务验证：至少需要提供一个字段
	if username == "" && email == "" && phone == "" {
		return ErrNoFieldProvided
	}

	// 检查用户名
	if username != "" {
		exists, err := s.userRepo.ExistsWithUsername(username)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCheckAvailability, err)
		}
		if exists {
			return ErrUsernameAlreadyExists
		}
	}

	// 检查邮箱
	if email != "" {
		exists, err := s.userRepo.ExistsWithEmail(email)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCheckAvailability, err)
		}
		if exists {
			return ErrEmailAlreadyExists
		}
	}

	// 检查手机号
	if phone != "" {
		exists, err := s.userRepo.ExistsWithPhone(phone)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCheckAvailability, err)
		}
		if exists {
			return ErrPhoneAlreadyExists
		}
	}

	return nil
}

// #endregion

// #region 用户列表和搜索

// GetUserList 获取用户列表（业务层增强 - 权限检查、数据过滤、缓存等）
func (s *UserService) GetUserList(offset, limit int) ([]*model.User, int64, error) {
	// 业务验证：分页参数检查
	if offset < 0 {
		return nil, 0, ErrPaginationInvalid
	}
	if limit <= 0 || limit > 100 { // 业务规则：限制最大页面大小
		return nil, 0, ErrPaginationInvalid
	}

	// 调用数据层
	users, total, err := s.userRepo.GetUserList(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// 业务层数据处理：移除敏感信息
	filteredUsers := make([]*model.User, len(users))
	for i, user := range users {
		filteredUser := *user
		// 清空敏感信息
		filteredUser.PasswordHash = ""
		filteredUsers[i] = &filteredUser
	}

	return filteredUsers, total, nil
}

// SearchUsers 搜索用户（业务层增强 - 搜索日志、结果过滤等）
func (s *UserService) SearchUsers(keyword string, offset, limit int) ([]*model.User, int64, error) {
	// 业务验证：分页参数检查
	if offset < 0 {
		return nil, 0, ErrPaginationInvalid
	}
	if limit <= 0 || limit > 100 {
		return nil, 0, ErrPaginationInvalid
	}

	// 业务验证：关键词长度检查
	if keyword != "" && len(keyword) < 2 {
		return nil, 0, ErrSearchKeywordShort
	}

	// 调用数据层
	users, total, err := s.userRepo.SearchUsers(keyword, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// 业务层数据处理：移除敏感信息
	filteredUsers := make([]*model.User, len(users))
	for i, user := range users {
		filteredUser := *user
		// 清空敏感信息
		filteredUser.PasswordHash = ""
		filteredUsers[i] = &filteredUser
	}

	// 业务日志：记录搜索操作（用于分析用户行为）
	if keyword != "" {
		log.Printf("用户搜索 - 关键词: %s, 结果数: %d, 时间: %s",
			keyword, total, time.Now().Format("2006-01-02 15:04:05"))
	}

	return filteredUsers, total, nil
}

// #endregion

// #region 用户统计

// GetUserStats 获取用户统计信息（业务层增强）
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总用户数
	totalUsers, err := s.userRepo.CountUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get total user count: %w", err)
	}
	stats["total_users"] = totalUsers

	// 活跃用户数（可以根据具体业务定义）
	activeUsers, err := s.userRepo.CountUsersByStatus("active")
	if err != nil {
		return nil, fmt.Errorf("failed to get active user count: %w", err)
	}
	stats["active_users"] = activeUsers

	// 业务层计算：用户增长率等更复杂的统计
	if totalUsers > 0 {
		stats["active_rate"] = float64(activeUsers) / float64(totalUsers) * 100
	} else {
		stats["active_rate"] = 0.0
	}

	return stats, nil
}

// #endregion

// #region 私有辅助方法

// getStoredSMSCode 从缓存或数据库获取手机号对应的验证码
func (s *UserService) getStoredSMSCode(phone string) (string, error) {
	// TODO: 实现从缓存或数据库获取验证码的逻辑
	// 例如：return cache.GetSMSCode(phone)
	// 这里暂时返回错误，实际部署时需实现
	return "", fmt.Errorf("getStoredSMSCode not implemented")
}

// getUserByLoginInfo 根据登录信息获取用户
func (s *UserService) getUserByLoginInfo(loginInfo string) (*model.User, error) {
	if validator.IsEmail(loginInfo) {
		return s.userRepo.GetUserByEmail(loginInfo)
	} else if validator.IsPhone(loginInfo) {
		return s.userRepo.GetUserByPhone(loginInfo)
	} else {
		return s.userRepo.GetUserByUsername(loginInfo)
	}
}

// verifyLoginCredentials 验证登录凭据
func (s *UserService) verifyLoginCredentials(user *model.User, loginInfo, password, loginType string) error {
	switch loginType {
	case "password":
		if err := crypto.VerifyPassword(user.PasswordHash, password); err != nil {
			return fmt.Errorf("invalid password")
		}
	case "sms":
		if !validator.IsPhone(loginInfo) {
			return fmt.Errorf("SMS login only supported for phone numbers")
		}
		// 从缓存或数据库获取实际验证码进行比较
		storedCode, err := s.getStoredSMSCode(loginInfo)
		if err != nil {
			return fmt.Errorf("failed to retrieve SMS code: %w", err)
		}
		if err := sms.Verify(password, storedCode); err != nil {
			return fmt.Errorf("invalid SMS code")
		}
	default:
		return fmt.Errorf("unsupported login type: %s", loginType)
	}

	return nil
}

// generateToken 生成JWT Token
func (s *UserService) generateToken(user *model.User) (string, error) {
	return s.jwtService.GenerateToken(user.ID, "user")
}

// validateUserFields 验证用户字段
func (s *UserService) validateUserFields(user *model.User) error {
	if user.Username != "" {
		if err := user.ValidateUsername(); err != nil {
			return err
		}
	}

	if user.Email != "" {
		if err := user.ValidateEmail(); err != nil {
			return err
		}
	}

	if user.Phone != "" {
		if err := user.ValidatePhone(); err != nil {
			return err
		}
	}

	return nil
}

// checkUserAvailabilityExcluding 检查用户可用性（排除指定用户）
func (s *UserService) checkUserAvailabilityExcluding(excludeUserID uint, username, email, phone string) error {
	// 检查用户名
	if username != "" {
		exists, err := s.userRepo.ExistsWithUsername(username)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCheckAvailability, err)
		}
		if exists {
			// 检查是否是当前用户
			existingUser, err := s.userRepo.GetUserByUsername(username)
			if err == nil && existingUser.ID != int64(excludeUserID) {
				return ErrUsernameAlreadyExists
			}
		}
	}

	// 检查邮箱
	if email != "" {
		exists, err := s.userRepo.ExistsWithEmail(email)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCheckAvailability, err)
		}
		if exists {
			existingUser, err := s.userRepo.GetUserByEmail(email)
			if err == nil && existingUser.ID != int64(excludeUserID) {
				return ErrEmailAlreadyExists
			}
		}
	}

	// 检查手机号
	if phone != "" {
		exists, err := s.userRepo.ExistsWithPhone(phone)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCheckAvailability, err)
		}
		if exists {
			existingUser, err := s.userRepo.GetUserByPhone(phone)
			if err == nil && existingUser.ID != int64(excludeUserID) {
				return ErrPhoneAlreadyExists
			}
		}
	}

	return nil
}

// checkSMSRateLimit 检查短信发送频率限制
func (s *UserService) checkSMSRateLimit(phone string) error {
	// 实现具体的频率限制逻辑
	// 这里可以检查数据库或缓存中是否有最近发送记录
	return nil
}

// #endregion

// #region 日志记录方法

// logUserRegistered 记录用户注册日志
func (s *UserService) logUserRegistered(user *model.User) {
	log.Printf("用户注册成功 - 用户名: %s, 邮箱: %s, 时间: %s",
		user.Username, user.MaskEmail(), time.Now().Format("2006-01-02 15:04:05"))
}

// logUserLogin 记录用户登录日志
func (s *UserService) logUserLogin(user *model.User, loginType string) {
	log.Printf("用户登录成功 - 用户ID: %d, 用户名: %s, 登录方式: %s, 时间: %s",
		user.ID, user.Username, loginType, time.Now().Format("2006-01-02 15:04:05"))
}

// logSMSSent 记录短信发送日志
func (s *UserService) logSMSSent(phone string, userID int64) {
	log.Printf("短信发送记录 - 手机号: %s, 用户ID: %d, 时间: %s",
		phone, userID, time.Now().Format("2006-01-02 15:04:05"))
}

// logUserProfileUpdated 记录用户档案更新日志
func (s *UserService) logUserProfileUpdated(userID uint) {
	log.Printf("用户档案更新 - 用户ID: %d, 时间: %s",
		userID, time.Now().Format("2006-01-02 15:04:05"))
}

// logPasswordUpdated 记录密码更新日志
func (s *UserService) logPasswordUpdated(userID uint) {
	log.Printf("用户密码更新 - 用户ID: %d, 时间: %s",
		userID, time.Now().Format("2006-01-02 15:04:05"))
}

// logPasswordReset 记录密码重置日志
func (s *UserService) logPasswordReset(userID int64) {
	log.Printf("用户密码重置 - 用户ID: %d, 时间: %s",
		userID, time.Now().Format("2006-01-02 15:04:05"))
}

// #endregion
