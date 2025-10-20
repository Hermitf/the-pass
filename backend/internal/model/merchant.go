package model

import (
	"gorm.io/gorm"
	"regexp"
	"time"

	"github.com/Hermitf/the-pass/pkg/formatting"
)

// #region 模型定义

// Merchant 商家模型
type Merchant struct {
	ID              int64          `json:"id" gorm:"primaryKey;autoIncrement;comment:商家ID"`
	Username        string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null;comment:用户名"`
	PasswordHash    string         `json:"-" gorm:"type:varchar(255);not null;comment:密码哈希"`
	Email           string         `json:"email" gorm:"type:varchar(100);uniqueIndex;not null;comment:邮箱"`
	Phone           string         `json:"phone" gorm:"type:varchar(20);uniqueIndex;not null;comment:手机号"`
	CompanyName     string         `json:"company_name" gorm:"type:varchar(100);comment:公司名称"`
	BusinessLicense string         `json:"business_license" gorm:"type:varchar(100);uniqueIndex;comment:营业执照号"`
	IsActive        bool           `json:"is_active" gorm:"default:true;comment:是否激活"`
	Employees       []Employee     `json:"employees,omitempty" gorm:"foreignKey:MerchantID"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 设置表名
func (Merchant) TableName() string {
	return "merchants"
}

// #endregion

// #region 响应DTO

// MerchantResponse 商家响应DTO（用于API返回）
type MerchantResponse struct {
	ID              int64  `json:"id"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	CompanyName     string `json:"company_name,omitempty"`
	BusinessLicense string `json:"business_license,omitempty"`
	IsActive        bool   `json:"is_active"`
}

// ToResponse 将 Merchant 模型转换为响应DTO
func (m *Merchant) ToResponse() *MerchantResponse {
	return &MerchantResponse{
		ID:              m.ID,
		Username:        m.Username,
		Email:           m.Email,
		Phone:           m.Phone,
		CompanyName:     m.CompanyName,
		BusinessLicense: m.BusinessLicense,
		IsActive:        m.IsActive,
	}
}

// MerchantSafeResponse 商家安全响应DTO（脱敏显示）
type MerchantSafeResponse struct {
	ID              int64  `json:"id"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	CompanyName     string `json:"company_name,omitempty"`
	BusinessLicense string `json:"business_license,omitempty"` // 脱敏显示
	IsActive        bool   `json:"is_active"`
	EmployeeCount   int    `json:"employee_count"` // 员工数量
}

// ToSafeResponse 转换为安全响应（脱敏）
func (m *Merchant) ToSafeResponse() *MerchantSafeResponse {
	return &MerchantSafeResponse{
		ID:              m.ID,
		Username:        m.Username,
		Email:           m.MaskEmail(),
		Phone:           m.MaskPhone(),
		CompanyName:     m.CompanyName,
		BusinessLicense: m.MaskBusinessLicense(),
		IsActive:        m.IsActive,
		EmployeeCount:   len(m.Employees),
	}
}

// #endregion

// #region 验证方法

// ValidateCompanyName 验证公司名称
func (m *Merchant) ValidateCompanyName() error {
	if m.CompanyName == "" {
		return ErrCompanyNameRequired
	}
	if len(m.CompanyName) > 100 {
		return ErrInvalidInput
	}
	return nil
}

// ValidateBusinessLicense 验证营业执照号
func (m *Merchant) ValidateBusinessLicense() error {
	if m.BusinessLicense == "" {
		return nil // 营业执照号可选
	}

	// 统一社会信用代码验证（18位）
	matched, _ := regexp.MatchString(`^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$`, m.BusinessLicense)
	if !matched {
		return ErrInvalidBusinessLicense
	}
	return nil
}

// ValidateAll 验证所有字段
func (m *Merchant) ValidateAll() error {
	if err := m.ValidateCompanyName(); err != nil {
		return err
	}
	if err := m.ValidateBusinessLicense(); err != nil {
		return err
	}
	return nil
}

// #endregion

// #region 业务方法

// IsActiveMerchant 检查商家是否为激活状态
func (m *Merchant) IsActiveMerchant() bool {
	return m.IsActive
}

// GetDisplayName 获取显示名称
func (m *Merchant) GetDisplayName() string {
	if m.CompanyName != "" {
		return m.CompanyName
	}
	return m.Username
}

// GetEmployeeCount 获取员工数量
func (m *Merchant) GetEmployeeCount() int {
	return len(m.Employees)
}

// HasEmployee 检查是否有指定员工
func (m *Merchant) HasEmployee(employeeID int64) bool {
	for _, employee := range m.Employees {
		if employee.ID == employeeID {
			return true
		}
	}
	return false
}

// GetActiveEmployees 获取激活的员工列表
func (m *Merchant) GetActiveEmployees() []Employee {
	var activeEmployees []Employee
	for _, employee := range m.Employees {
		if employee.IsActive {
			activeEmployees = append(activeEmployees, employee)
		}
	}
	return activeEmployees
}

// UpdateProfile 更新商家基本信息
func (m *Merchant) UpdateProfile(companyName, businessLicense string) error {
	// 创建临时对象进行验证
	tempMerchant := &Merchant{
		CompanyName:     companyName,
		BusinessLicense: businessLicense,
	}

	if err := tempMerchant.ValidateAll(); err != nil {
		return err
	}

	m.CompanyName = companyName
	m.BusinessLicense = businessLicense
	m.UpdatedAt = time.Now()

	return nil
}

// Activate 激活商家账号
func (m *Merchant) Activate() {
	m.IsActive = true
	m.UpdatedAt = time.Now()
}

// Deactivate 停用商家账号
func (m *Merchant) Deactivate() {
	m.IsActive = false
	m.UpdatedAt = time.Now()
}

// AddEmployee 添加员工（逻辑关联，实际添加在service层处理）
func (m *Merchant) AddEmployee(employee *Employee) {
	employee.MerchantID = m.ID
	m.Employees = append(m.Employees, *employee)
}

// RemoveEmployee 移除员工（逻辑移除，实际删除在service层处理）
func (m *Merchant) RemoveEmployee(employeeID int64) {
	for i, employee := range m.Employees {
		if employee.ID == employeeID {
			m.Employees = append(m.Employees[:i], m.Employees[i+1:]...)
			break
		}
	}
}

// #endregion

// #region 工具方法

// MaskEmail 脱敏显示邮箱
func (m *Merchant) MaskEmail() string {
	return formatting.MaskEmail(m.Email)
}

// MaskPhone 脱敏显示手机号
func (m *Merchant) MaskPhone() string {
	return formatting.MaskPhone(m.Phone)
}

// MaskBusinessLicense 脱敏显示营业执照号
func (m *Merchant) MaskBusinessLicense() string {
	if m.BusinessLicense == "" {
		return ""
	}

	if len(m.BusinessLicense) >= 10 {
		return m.BusinessLicense[:6] + "******" + m.BusinessLicense[len(m.BusinessLicense)-4:]
	}
	return m.BusinessLicense
}

// GetBusinessInfo 获取商家业务信息摘要
func (m *Merchant) GetBusinessInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":               m.ID,
		"company_name":     m.CompanyName,
		"business_license": m.MaskBusinessLicense(),
		"employee_count":   m.GetEmployeeCount(),
		"active_employees": len(m.GetActiveEmployees()),
		"is_active":        m.IsActive,
		"created_at":       m.CreatedAt,
	}
}

// #endregion
