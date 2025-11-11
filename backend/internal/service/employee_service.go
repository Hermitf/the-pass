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
const logTimeLayout = "2006-01-02 15:04:05"

// EmployeeServiceInterface 员工服务接口
type EmployeeServiceInterface interface {
	// 员工注册和认证
	RegisterEmployee(employee *model.Employee) error
	LoginEmployee(loginInfo, password, loginType string) (string, error)

	// 员工信息管理
	GetEmployeeByID(id int64) (*model.Employee, error)
	UpdateEmployeeProfile(employeeID int64, name, email, phone string) error
	UpdateEmployeePassword(employeeID int64, oldPassword, newPassword string) error

	// 商家关联管理
	GetEmployeesByMerchantID(merchantID int64) ([]*model.Employee, error)
	GetActiveEmployeesByMerchant(merchantID int64) ([]*model.Employee, error)
	TransferEmployee(employeeID, newMerchantID int64) error

	// 员工验证
	ValidateEmployeeData(employee *model.Employee) error
	CheckEmployeeAvailability(username, email, phone string) error

	// 员工列表和搜索
	GetEmployeeList(merchantID int64, offset, limit int) ([]*model.Employee, int64, error)
	SearchEmployees(keyword string, merchantID int64, offset, limit int) ([]*model.Employee, int64, error)

	// 员工统计
	GetEmployeeStatsByMerchant(merchantID int64) (map[string]interface{}, error)
}

// EmployeeService 员工服务实现
type EmployeeService struct {
	employeeRepo repository.EmployeeRepositoryInterface
	jwtService   JWTServiceInterface
}

// #endregion

// #region 构造函数和依赖注入

// EmployeeServiceDependencies 员工服务依赖
type EmployeeServiceDependencies struct {
	EmployeeRepo repository.EmployeeRepositoryInterface
	JWTService   JWTServiceInterface
}

// NewEmployeeService 创建员工服务实例
func NewEmployeeService(deps EmployeeServiceDependencies) EmployeeServiceInterface {
	return &EmployeeService{
		employeeRepo: deps.EmployeeRepo,
		jwtService:   deps.JWTService,
	}
}

// #endregion

// #region 员工注册和认证

// RegisterEmployee 注册员工
func (s *EmployeeService) RegisterEmployee(employee *model.Employee) error {
	if employee == nil {
		return ErrEmployeeNil
	}

	// 验证员工数据
	if err := s.ValidateEmployeeData(employee); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 检查员工是否已存在
	if err := s.CheckEmployeeAvailability(employee.Username, employee.Email, employee.Phone); err != nil {
		return fmt.Errorf("%w: %v", ErrAvailabilityCheck, err)
	}

	// 加密密码
	if employee.PasswordHash != "" {
		hashedPassword, err := crypto.HashPassword(employee.PasswordHash)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPasswordHashing, err)
		}
		employee.PasswordHash = hashedPassword
	}

	// 创建员工
	if err := s.employeeRepo.Create(employee); err != nil {
		return fmt.Errorf("%w: %v", ErrDataSaveFailed, err)
	}

	s.logEmployeeRegistered(employee)
	return nil
}

// LoginEmployee 员工登录
func (s *EmployeeService) LoginEmployee(loginInfo, password, loginType string) (string, error) {
	if loginInfo == "" || password == "" {
		return "", ErrLoginInfoEmpty
	}

	// 根据登录类型获取员工信息
	employee, err := s.getEmployeeByLoginInfo(loginInfo, loginType)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEmployeeNotFound, err)
	}

	// 验证密码
	if err := crypto.VerifyPassword(employee.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	// 检查员工状态
	if !employee.IsActive {
		return "", ErrAccountDeactivated
	}

	// 生成JWT令牌
	token, err := s.jwtService.GenerateToken(employee.ID, "employee")
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	s.logEmployeeLogin(employee, loginType)
	return token, nil
}

// #endregion

// #region 员工信息管理

// GetEmployeeByID 根据ID获取员工信息
func (s *EmployeeService) GetEmployeeByID(id int64) (*model.Employee, error) {
	return s.fetchEmployeeByID(id)
}

// UpdateEmployeeProfile 更新员工档案
func (s *EmployeeService) UpdateEmployeeProfile(employeeID int64, name, email, phone string) error {
	if employeeID <= 0 {
		return ErrInvalidEmployeeID
	}

	// 获取员工信息
	employee, err := s.fetchEmployeeByID(employeeID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEmployeeNotFound, err)
	}

	// 验证新的员工信息
	if err := s.validateEmployeeFields(email, phone); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 检查重复数据（排除当前员工）
	if err := s.checkEmployeeAvailabilityExcluding(employeeID, "", email, phone); err != nil {
		return err
	}

	// 更新员工信息
	employee.Name = name
	employee.Email = email
	employee.Phone = phone

	if err := s.employeeRepo.Update(employee); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logEmployeeProfileUpdated(employeeID)
	return nil
}

// UpdateEmployeePassword 更新员工密码
func (s *EmployeeService) UpdateEmployeePassword(employeeID int64, oldPassword, newPassword string) error {
	if employeeID <= 0 {
		return ErrInvalidEmployeeID
	}
	if oldPassword == "" || newPassword == "" {
		return ErrPasswordsEmpty
	}

	// 获取员工信息
	employee, err := s.fetchEmployeeByID(employeeID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := crypto.VerifyPassword(employee.PasswordHash, oldPassword); err != nil {
		return ErrOldPasswordIncorrect
	}

	// 加密新密码
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPasswordHashing, err)
	}

	// 更新密码
	employee.PasswordHash = hashedPassword
	if err := s.employeeRepo.Update(employee); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logEmployeePasswordUpdated(employeeID)
	return nil
}

// #endregion

// #region 商家关联管理

// GetEmployeesByMerchantID 根据商家ID获取员工列表
func (s *EmployeeService) GetEmployeesByMerchantID(merchantID int64) ([]*model.Employee, error) {
	if merchantID <= 0 {
		return nil, ErrInvalidMerchantID
	}

	employees, err := s.employeeRepo.GetByMerchantID(merchantID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetEmployeeList, err)
	}

	return employees, nil
}

// GetActiveEmployeesByMerchant 获取商家的活跃员工
func (s *EmployeeService) GetActiveEmployeesByMerchant(merchantID int64) ([]*model.Employee, error) {
	if merchantID <= 0 {
		return nil, ErrInvalidMerchantID
	}

	employees, err := s.employeeRepo.GetActiveEmployeesByMerchant(merchantID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetEmployeeList, err)
	}

	return employees, nil
}

// TransferEmployee 转移员工到新商家
func (s *EmployeeService) TransferEmployee(employeeID, newMerchantID int64) error {
	if employeeID <= 0 || newMerchantID <= 0 {
		return ErrInvalidEmployeeID
	}

	// 获取员工信息
	employee, err := s.fetchEmployeeByID(employeeID)
	if err != nil {
		return err
	}

	// 检查是否转移到相同商家
	if employee.MerchantID == newMerchantID {
		return ErrSameMerchantTransfer
	}

	// 执行转移
	if err := s.employeeRepo.TransferEmployee(employeeID, newMerchantID); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logEmployeeTransferred(employeeID, employee.MerchantID, newMerchantID)
	return nil
}

// #endregion

// #region 员工验证

// ValidateEmployeeData 验证员工数据
func (s *EmployeeService) ValidateEmployeeData(employee *model.Employee) error {
	if employee == nil {
		return ErrEmployeeNil
	}

	return s.validateEmployeeFields(employee.Email, employee.Phone)
}

// CheckEmployeeAvailability 检查员工可用性
func (s *EmployeeService) CheckEmployeeAvailability(username, email, phone string) error {
	exists, err := s.employeeRepo.CheckEmployeeExists(username, email, phone)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAvailabilityCheck, err)
	}

	if exists {
		return ErrEmployeeAlreadyExists
	}

	return nil
}

// #endregion

// #region 员工列表和搜索

// GetEmployeeList 获取员工列表
func (s *EmployeeService) GetEmployeeList(merchantID int64, offset, limit int) ([]*model.Employee, int64, error) {
	if merchantID <= 0 {
		return nil, 0, ErrInvalidMerchantID
	}
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	employees, total, err := s.employeeRepo.GetEmployeesByMerchantWithPagination(merchantID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrGetEmployeeList, err)
	}

	return employees, total, nil
}

// SearchEmployees 搜索员工
func (s *EmployeeService) SearchEmployees(keyword string, merchantID int64, offset, limit int) ([]*model.Employee, int64, error) {
	if merchantID <= 0 {
		return nil, 0, ErrInvalidMerchantID
	}
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	employees, total, err := s.employeeRepo.SearchEmployees(keyword, merchantID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrSearchEmployees, err)
	}

	return employees, total, nil
}

// #endregion

// #region 员工统计

// GetEmployeeStatsByMerchant 获取商家员工统计信息
func (s *EmployeeService) GetEmployeeStatsByMerchant(merchantID int64) (map[string]interface{}, error) {
	if merchantID <= 0 {
		return nil, ErrInvalidMerchantID
	}

	stats, err := s.employeeRepo.GetEmployeeStatsByMerchant(merchantID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetStatistics, err)
	}

	return stats, nil
}

// #endregion

// #region 私有辅助方法
func (s *EmployeeService) fetchEmployeeByID(id int64) (*model.Employee, error) {
	if id <= 0 {
		return nil, ErrInvalidEmployeeID
	}

	employee, err := s.employeeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEmployeeNotFound, err)
	}

	return employee, nil
}

func (s *EmployeeService) now() string {
	return time.Now().Format(logTimeLayout)
}

// getEmployeeByLoginInfo 根据登录信息获取员工
func (s *EmployeeService) getEmployeeByLoginInfo(loginInfo, loginType string) (*model.Employee, error) {
	switch loginType {
	case "email":
		return s.employeeRepo.GetByEmail(loginInfo)
	case "phone":
		return s.employeeRepo.GetByPhone(loginInfo)
	default:
		// 智能检测登录类型
		if validator.IsEmail(loginInfo) {
			return s.employeeRepo.GetByEmail(loginInfo)
		} else if validator.IsPhone(loginInfo) {
			return s.employeeRepo.GetByPhone(loginInfo)
		} else {
			return s.employeeRepo.GetByUsername(loginInfo)
		}
	}
}

// validateEmployeeFields 验证员工字段
func (s *EmployeeService) validateEmployeeFields(email, phone string) error {
	if email != "" && !validator.IsEmail(email) {
		return ErrEmailInvalid
	}

	if phone != "" && !validator.IsPhone(phone) {
		return ErrPhoneInvalid
	}

	return nil
}

// checkEmployeeAvailabilityExcluding 检查员工可用性（排除指定员工）
func (s *EmployeeService) checkEmployeeAvailabilityExcluding(excludeEmployeeID int64, username, email, phone string) error {
	// 检查邮箱
	if email != "" {
		existing, err := s.employeeRepo.GetByEmail(email)
		if err == nil && existing.ID != excludeEmployeeID {
			return ErrEmailAlreadyExists
		}
	}

	// 检查手机号
	if phone != "" {
		existing, err := s.employeeRepo.GetByPhone(phone)
		if err == nil && existing.ID != excludeEmployeeID {
			return ErrPhoneAlreadyExists
		}
	}

	// 检查用户名
	if username != "" {
		existing, err := s.employeeRepo.GetByUsername(username)
		if err == nil && existing.ID != excludeEmployeeID {
			return ErrUsernameAlreadyExists
		}
	}

	return nil
}

// #endregion

// #region 日志记录方法

// logEmployeeRegistered 记录员工注册日志
func (s *EmployeeService) logEmployeeRegistered(employee *model.Employee) {
	log.Printf("员工注册成功 - 用户名: %s, 邮箱: %s, 商家ID: %d, 时间: %s",
		employee.Username, employee.Email, employee.MerchantID, s.now())
}

// logEmployeeLogin 记录员工登录日志
func (s *EmployeeService) logEmployeeLogin(employee *model.Employee, loginType string) {
	log.Printf("员工登录成功 - 员工ID: %d, 用户名: %s, 商家ID: %d, 登录方式: %s, 时间: %s",
		employee.ID, employee.Username, employee.MerchantID, loginType, s.now())
}

// logEmployeeProfileUpdated 记录员工档案更新日志
func (s *EmployeeService) logEmployeeProfileUpdated(employeeID int64) {
	log.Printf("员工档案更新 - 员工ID: %d, 时间: %s",
		employeeID, s.now())
}

// logEmployeePasswordUpdated 记录员工密码更新日志
func (s *EmployeeService) logEmployeePasswordUpdated(employeeID int64) {
	log.Printf("员工密码更新 - 员工ID: %d, 时间: %s",
		employeeID, s.now())
}

// logEmployeeTransferred 记录员工转移日志
func (s *EmployeeService) logEmployeeTransferred(employeeID, oldMerchantID, newMerchantID int64) {
	log.Printf("员工转移 - 员工ID: %d, 原商家ID: %d, 新商家ID: %d, 时间: %s",
		employeeID, oldMerchantID, newMerchantID, s.now())
}

// #endregion
