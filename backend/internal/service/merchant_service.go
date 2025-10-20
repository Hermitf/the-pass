package service

import (
	"fmt"
	"log"
	"time"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/repository"
	"github.com/Hermitf/the-pass/pkg/crypto"
	"github.com/Hermitf/the-pass/pkg/validator"
)

// #region 服务定义

// MerchantServiceInterface 商家服务接口
type MerchantServiceInterface interface {
	// 商家注册和认证
	RegisterMerchant(merchant *model.Merchant) error
	LoginMerchant(loginInfo, password, loginType string) (string, error)

	// 商家信息管理
	GetMerchantByID(id int64) (*model.Merchant, error)
	UpdateMerchantProfile(merchantID int64, companyName, address, contactName string) error
	UpdateMerchantPassword(merchantID int64, oldPassword, newPassword string) error

	// 商家验证
	ValidateMerchantData(merchant *model.Merchant) error
	CheckMerchantAvailability(username, email, phone, businessLicense string) error

	// 员工管理
	GetMerchantWithEmployees(merchantID int64) (*model.Merchant, []*model.Employee, error)
	GetMerchantEmployeeStats(merchantID int64) (map[string]interface{}, error)

	// 商家列表和搜索
	GetMerchantList(offset, limit int) ([]*model.Merchant, int64, error)
	GetActiveMerchants(offset, limit int) ([]*model.Merchant, int64, error)
	SearchMerchants(keyword string, offset, limit int) ([]*model.Merchant, int64, error)
	GetMerchantsByRegion(region string, offset, limit int) ([]*model.Merchant, int64, error)

	// 商家统计
	GetMerchantStats() (map[string]interface{}, error)
	GetTopMerchantsByEmployeeCount(limit int) ([]*model.Merchant, error)
}

// MerchantService 商家服务实现
type MerchantService struct {
	merchantRepo repository.MerchantRepositoryInterface
	employeeRepo repository.EmployeeRepositoryInterface
	jwtService   JWTServiceInterface
}

// #endregion

// #region 构造函数和依赖注入

// MerchantServiceDependencies 商家服务依赖
type MerchantServiceDependencies struct {
	MerchantRepo repository.MerchantRepositoryInterface
	EmployeeRepo repository.EmployeeRepositoryInterface
	JWTService   JWTServiceInterface
}

// NewMerchantService 创建商家服务实例
func NewMerchantService(deps MerchantServiceDependencies) MerchantServiceInterface {
	return &MerchantService{
		merchantRepo: deps.MerchantRepo,
		employeeRepo: deps.EmployeeRepo,
		jwtService:   deps.JWTService,
	}
}

// #endregion

// #region 商家注册和认证

// RegisterMerchant 注册商家
func (s *MerchantService) RegisterMerchant(merchant *model.Merchant) error {
	if merchant == nil {
		return ErrMerchantNil
	}

	// 验证商家数据
	if err := s.ValidateMerchantData(merchant); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 检查商家是否已存在
	if err := s.CheckMerchantAvailability(merchant.Username, merchant.Email, merchant.Phone, merchant.BusinessLicense); err != nil {
		return fmt.Errorf("%w: %v", ErrAvailabilityCheck, err)
	}

	// 加密密码
	if merchant.PasswordHash != "" {
		hashedPassword, err := crypto.HashPassword(merchant.PasswordHash)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPasswordHashing, err)
		}
		merchant.PasswordHash = hashedPassword
	}

	// 创建商家
	if err := s.merchantRepo.Create(merchant); err != nil {
		return fmt.Errorf("%w: %v", ErrDataSaveFailed, err)
	}

	s.logMerchantRegistered(merchant)
	return nil
}

// LoginMerchant 商家登录
func (s *MerchantService) LoginMerchant(loginInfo, password, loginType string) (string, error) {
	if loginInfo == "" || password == "" {
		return "", ErrLoginInfoEmpty
	}

	// 根据登录类型获取商家信息
	merchant, err := s.getMerchantByLoginInfo(loginInfo, loginType)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrMerchantNotFound, err)
	}

	// 验证密码
	if err := crypto.VerifyPassword(merchant.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	// 检查商家状态
	if !merchant.IsActive {
		return "", ErrAccountDeactivated
	}

	// 生成JWT令牌
	token, err := s.jwtService.GenerateToken(merchant.ID, "merchant")
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	s.logMerchantLogin(merchant, loginType)
	return token, nil
}

// #endregion

// #region 商家信息管理

// GetMerchantByID 根据ID获取商家信息
func (s *MerchantService) GetMerchantByID(id int64) (*model.Merchant, error) {
	if id <= 0 {
		return nil, ErrInvalidMerchantID
	}

	merchant, err := s.merchantRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMerchantNotFound, err)
	}

	return merchant, nil
}

// UpdateMerchantProfile 更新商家档案
func (s *MerchantService) UpdateMerchantProfile(merchantID int64, companyName, address, contactName string) error {
	if merchantID <= 0 {
		return ErrInvalidMerchantID
	}

	// 获取商家信息
	merchant, err := s.merchantRepo.GetByID(merchantID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMerchantNotFound, err)
	}

	// 验证新的商家信息
	if err := s.validateMerchantFields(companyName); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 更新商家信息
	merchant.CompanyName = companyName
	// 注意：当前模型没有Address和ContactName字段，只更新CompanyName

	if err := s.merchantRepo.Update(merchant); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logMerchantProfileUpdated(merchantID)
	return nil
}

// UpdateMerchantPassword 更新商家密码
func (s *MerchantService) UpdateMerchantPassword(merchantID int64, oldPassword, newPassword string) error {
	if merchantID <= 0 {
		return ErrInvalidMerchantID
	}
	if oldPassword == "" || newPassword == "" {
		return ErrPasswordsEmpty
	}

	// 获取商家信息
	merchant, err := s.merchantRepo.GetByID(merchantID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMerchantNotFound, err)
	}

	// 验证旧密码
	if err := crypto.VerifyPassword(merchant.PasswordHash, oldPassword); err != nil {
		return ErrOldPasswordIncorrect
	}

	// 加密新密码
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPasswordHashing, err)
	}

	// 更新密码
	merchant.PasswordHash = hashedPassword
	if err := s.merchantRepo.Update(merchant); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logMerchantPasswordUpdated(merchantID)
	return nil
}

// #endregion

// #region 商家验证

// ValidateMerchantData 验证商家数据
func (s *MerchantService) ValidateMerchantData(merchant *model.Merchant) error {
	if merchant == nil {
		return ErrMerchantNil
	}

	return s.validateMerchantFields(merchant.CompanyName)
}

// CheckMerchantAvailability 检查商家可用性
func (s *MerchantService) CheckMerchantAvailability(username, email, phone, businessLicense string) error {
	exists, err := s.merchantRepo.CheckMerchantExists(username, email, phone, businessLicense)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAvailabilityCheck, err)
	}

	if exists {
		return ErrMerchantExists
	}

	return nil
}

// #endregion

// #region 员工管理

// GetMerchantWithEmployees 获取商家及其员工信息
func (s *MerchantService) GetMerchantWithEmployees(merchantID int64) (*model.Merchant, []*model.Employee, error) {
	if merchantID <= 0 {
		return nil, nil, ErrInvalidMerchantID
	}

	merchant, employees, err := s.merchantRepo.GetMerchantWithEmployees(merchantID)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrMerchantNotFound, err)
	}

	return merchant, employees, nil
}

// GetMerchantEmployeeStats 获取商家员工统计信息
func (s *MerchantService) GetMerchantEmployeeStats(merchantID int64) (map[string]interface{}, error) {
	if merchantID <= 0 {
		return nil, ErrInvalidMerchantID
	}

	if s.employeeRepo == nil {
		return nil, ErrEmployeeRepoUnavailable
	}

	stats, err := s.employeeRepo.GetEmployeeStatsByMerchant(merchantID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetStatistics, err)
	}

	return stats, nil
}

// #endregion

// #region 商家列表和搜索

// GetMerchantList 获取商家列表
func (s *MerchantService) GetMerchantList(offset, limit int) ([]*model.Merchant, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	merchants, total, err := s.merchantRepo.GetMerchantList(offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrGetMerchantList, err)
	}

	return merchants, total, nil
}

// GetActiveMerchants 获取活跃商家列表
func (s *MerchantService) GetActiveMerchants(offset, limit int) ([]*model.Merchant, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	merchants, total, err := s.merchantRepo.GetActiveMerchants(offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrGetMerchantList, err)
	}

	return merchants, total, nil
}

// SearchMerchants 搜索商家
func (s *MerchantService) SearchMerchants(keyword string, offset, limit int) ([]*model.Merchant, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	merchants, total, err := s.merchantRepo.SearchMerchants(keyword, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrSearchMerchants, err)
	}

	return merchants, total, nil
}

// GetMerchantsByRegion 根据地区获取商家
func (s *MerchantService) GetMerchantsByRegion(region string, offset, limit int) ([]*model.Merchant, int64, error) {
	if region == "" {
		return nil, 0, ErrRegionEmpty
	}
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	merchants, total, err := s.merchantRepo.GetMerchantsByRegion(region, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrGetMerchantList, err)
	}

	return merchants, total, nil
}

// #endregion

// #region 商家统计

// GetMerchantStats 获取商家统计信息
func (s *MerchantService) GetMerchantStats() (map[string]interface{}, error) {
	stats, err := s.merchantRepo.GetMerchantStats()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetStatistics, err)
	}

	return stats, nil
}

// GetTopMerchantsByEmployeeCount 获取员工数量最多的商家
func (s *MerchantService) GetTopMerchantsByEmployeeCount(limit int) ([]*model.Merchant, error) {
	if limit <= 0 {
		return nil, ErrLimitInvalid
	}

	merchants, err := s.merchantRepo.GetTopMerchantsByEmployeeCount(limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetMerchantList, err)
	}

	return merchants, nil
}

// #endregion

// #region 私有辅助方法

// getMerchantByLoginInfo 根据登录信息获取商家
func (s *MerchantService) getMerchantByLoginInfo(loginInfo, loginType string) (*model.Merchant, error) {
	switch loginType {
	case "email":
		return s.merchantRepo.GetByEmail(loginInfo)
	case "phone":
		return s.merchantRepo.GetByPhone(loginInfo)
	default:
		// 智能检测登录类型
		if validator.IsEmail(loginInfo) {
			return s.merchantRepo.GetByEmail(loginInfo)
		} else if validator.IsPhone(loginInfo) {
			return s.merchantRepo.GetByPhone(loginInfo)
		} else {
			return s.merchantRepo.GetByUsername(loginInfo)
		}
	}
}

// validateMerchantFields 验证商家字段
func (s *MerchantService) validateMerchantFields(companyName string) error {
	if companyName == "" {
		return ErrCompanyNameEmpty
	}

	if len(companyName) > 100 {
		return ErrCompanyNameTooLong
	}

	return nil
}

// #endregion

// #region 日志记录方法

// logMerchantRegistered 记录商家注册日志
func (s *MerchantService) logMerchantRegistered(merchant *model.Merchant) {
	log.Printf("商家注册成功 - 用户名: %s, 公司名: %s, 邮箱: %s, 时间: %s",
		merchant.Username, merchant.CompanyName, merchant.Email, time.Now().Format("2006-01-02 15:04:05"))
}

// logMerchantLogin 记录商家登录日志
func (s *MerchantService) logMerchantLogin(merchant *model.Merchant, loginType string) {
	log.Printf("商家登录成功 - 商家ID: %d, 用户名: %s, 公司名: %s, 登录方式: %s, 时间: %s",
		merchant.ID, merchant.Username, merchant.CompanyName, loginType, time.Now().Format("2006-01-02 15:04:05"))
}

// logMerchantProfileUpdated 记录商家档案更新日志
func (s *MerchantService) logMerchantProfileUpdated(merchantID int64) {
	log.Printf("商家档案更新 - 商家ID: %d, 时间: %s",
		merchantID, time.Now().Format("2006-01-02 15:04:05"))
}

// logMerchantPasswordUpdated 记录商家密码更新日志
func (s *MerchantService) logMerchantPasswordUpdated(merchantID int64) {
	log.Printf("商家密码更新 - 商家ID: %d, 时间: %s",
		merchantID, time.Now().Format("2006-01-02 15:04:05"))
}

// #endregion
