package service

import (
	"errors"
	"log"
	"time"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/repository"
	"github.com/Hermitf/the-pass/internal/utils/userutils"
)

// 用户服务
type UserService struct {
	userRepo   *repository.UserRepository
	jwtService *JWTService
}

// 创建用户服务
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo, jwtService: NewJWTService()}
}

// 注册用户
func (s *UserService) RegisterUser(user *model.User) error {
	if err := s.userRepo.CreateUser(user); err != nil {
		return err
	}
	return nil
}

// 统一登录函数
func (s *UserService) LoginUser(loginInfo string, password string, loginType string) (string, error) {
	var user *model.User
	var err error

	// 设置默认登录类型
	if loginType == "" {
		loginType = "password"
	}

	// 判断登录方式
	if userutils.IsEmail(loginInfo) {
		// 邮箱登录
		user, err = s.userRepo.GetUserByEmail(loginInfo)
		if err != nil {
			return "", err
		}

		// 邮箱只支持密码登录
		if !userutils.VerifyPassword(password, user.PasswordHash) {
			return "", errors.New("密码错误")
		}
	} else if userutils.IsPhone(loginInfo) {
		user, err = s.userRepo.GetByPhone(loginInfo)
		if err != nil {
			return "", err
		}
		switch loginType {
		case "password":
			// 手机号密码登录
			if !userutils.VerifyPassword(password, user.PasswordHash) {
				return "", errors.New("密码错误")
			}
		case "sms":
			// 手机号验证码登录
			if err := userutils.VerifySMS(loginInfo, password); err != nil {
				return "", errors.New("验证码错误")
			}
		default:
			return "", errors.New("无效的登录方式")
		}
	} else {
		// 用户名只支持密码登录
		user, err = s.userRepo.GetUserByUsername(loginInfo)
		if err != nil {
			return "", err
		}

		if !userutils.VerifyPassword(password, user.PasswordHash) {
			return "", errors.New("密码错误")
		}
	}

	// 生成 JWT Token
	token, err := s.generateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

// 发送短信验证码
func (s *UserService) SendSMSCode(phone string) error {
	// 1. 验证手机号格式
	if !userutils.IsPhone(phone) {
		return errors.New("手机号格式不正确")
	}

	// 2. 检查用户是否存在
	user, err := s.userRepo.GetByPhone(phone)
	if err != nil {
		return errors.New("手机号未注册")
	}

	// 3. 检查发送频率限制
	if err := s.checkSMSRateLimit(phone); err != nil {
		return err
	}

	// 4. 生成并保存验证码
	code := userutils.GenerateRandomCode()
	if err := userutils.SaveSMSCode(phone, code, 5*time.Minute); err != nil {
		return err
	}

	// 5. 发送短信
	if err := userutils.SendSMS(phone, code); err != nil {
		return err
	}

	// 6. 记录发送日志
	s.logSMSSent(phone, user.ID)

	return nil
}

// 生成JWT Token（使用JWT服务）
func (s *UserService) generateToken(user *model.User) (string, error) {
	return s.jwtService.GenerateUserToken(user)
}

// 检查短信发送频率限制
func (s *UserService) checkSMSRateLimit(phone string) error {
	// 这里可以实现具体的频率限制逻辑，比如检查数据库或缓存中是否有最近发送记录
	// 如果超过限制，返回错误
	return nil // 假设没有超过限制
}

// 记录短信发送日志
func (s *UserService) logSMSSent(phone string, userID int64) {
	// 这里可以实现具体的日志记录逻辑，比如写入数据库或日志文件
	// 记录发送时间、手机号、用户ID等信息
	log.Printf("短信发送记录 - 手机号: %s, 用户ID: %d, 时间: %s",
		phone, userID, time.Now().Format("2006-01-02 15:04:05"))
}
