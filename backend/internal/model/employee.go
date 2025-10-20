package model

import (
	"regexp"
	"time"

	"github.com/Hermitf/the-pass/pkg/formatting"
	"gorm.io/gorm"
)

// #region 模型定义

// Employee 员工模型
type Employee struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement;comment:员工ID"`
	Username     string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null;comment:用户名"`
	PasswordHash string         `json:"-" gorm:"type:varchar(255);not null;comment:密码哈希"`
	Email        string         `json:"email" gorm:"type:varchar(100);uniqueIndex;not null;comment:邮箱"`
	Phone        string         `json:"phone" gorm:"type:varchar(20);uniqueIndex;not null;comment:手机号"`
	Name         string         `json:"name" gorm:"type:varchar(50);comment:员工姓名"`
	IDNumber     string         `json:"id_number" gorm:"type:varchar(20);uniqueIndex;comment:身份证号"`
	Sex          string         `json:"sex" gorm:"type:varchar(10);comment:性别"`
	MerchantID   int64          `json:"merchant_id" gorm:"not null;index;comment:所属商家ID"`
	Merchant     *Merchant      `json:"merchant,omitempty" gorm:"foreignKey:MerchantID"`
	IsActive     bool           `json:"is_active" gorm:"default:true;comment:是否激活"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 设置表名
func (Employee) TableName() string {
	return "employees"
}

// #endregion

// #region 响应DTO

// EmployeeResponse 员工响应DTO（用于API返回）
type EmployeeResponse struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Name       string `json:"name,omitempty"`
	Sex        string `json:"sex,omitempty"`
	MerchantID int64  `json:"merchant_id"`
	IsActive   bool   `json:"is_active"`
}

// ToResponse 将 Employee 模型转换为响应DTO
func (e *Employee) ToResponse() *EmployeeResponse {
	return &EmployeeResponse{
		ID:         e.ID,
		Username:   e.Username,
		Email:      e.Email,
		Phone:      e.Phone,
		Name:       e.Name,
		Sex:        e.Sex,
		MerchantID: e.MerchantID,
		IsActive:   e.IsActive,
	}
}

// EmployeeSafeResponse 员工安全响应DTO（脱敏显示）
type EmployeeSafeResponse struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"` // 脱敏显示
	Phone      string `json:"phone"` // 脱敏显示
	Name       string `json:"name,omitempty"`
	Sex        string `json:"sex,omitempty"`
	MerchantID int64  `json:"merchant_id"`
	IsActive   bool   `json:"is_active"`
}

// ToSafeResponse 转换为安全响应（脱敏）
func (e *Employee) ToSafeResponse() *EmployeeSafeResponse {
	return &EmployeeSafeResponse{
		ID:         e.ID,
		Username:   e.Username,
		Email:      e.MaskEmail(),
		Phone:      e.MaskPhone(),
		Name:       e.Name,
		Sex:        e.Sex,
		MerchantID: e.MerchantID,
		IsActive:   e.IsActive,
	}
}

// #endregion

// #region 验证方法

// ValidateIDNumber 验证身份证号格式
func (e *Employee) ValidateIDNumber() error {
	if e.IDNumber == "" {
		return nil // 身份证号可选
	}

	// 18位身份证号验证
	matched, _ := regexp.MatchString(`^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`, e.IDNumber)
	if !matched {
		return ErrInvalidIDNumber
	}
	return nil
}

// ValidateSex 验证性别
func (e *Employee) ValidateSex() error {
	if e.Sex != "" && e.Sex != "男" && e.Sex != "女" && e.Sex != "male" && e.Sex != "female" {
		return ErrInvalidInput
	}
	return nil
}

// ValidateMerchantID 验证商家ID
func (e *Employee) ValidateMerchantID() error {
	if e.MerchantID <= 0 {
		return ErrInvalidMerchantID
	}
	return nil
}

// ValidateAll 验证所有字段
func (e *Employee) ValidateAll() error {
	if err := e.ValidateIDNumber(); err != nil {
		return err
	}
	if err := e.ValidateSex(); err != nil {
		return err
	}
	if err := e.ValidateMerchantID(); err != nil {
		return err
	}
	return nil
}

// #endregion

// #region 业务方法

// BelongsToMerchant 检查员工是否属于指定商家
func (e *Employee) BelongsToMerchant(merchantID int64) bool {
	return e.MerchantID == merchantID
}

// IsActiveEmployee 检查员工是否为激活状态
func (e *Employee) IsActiveEmployee() bool {
	return e.IsActive
}

// GetDisplayName 获取显示名称
func (e *Employee) GetDisplayName() string {
	if e.Name != "" {
		return e.Name
	}
	return e.Username
}

// UpdateProfile 更新员工基本信息
func (e *Employee) UpdateProfile(name, sex, idNumber string) error {
	// 创建临时对象进行验证
	tempEmployee := &Employee{
		Name:     name,
		Sex:      sex,
		IDNumber: idNumber,
	}

	if err := tempEmployee.ValidateAll(); err != nil {
		return err
	}

	e.Name = name
	e.Sex = sex
	e.IDNumber = idNumber
	e.UpdatedAt = time.Now()

	return nil
}

// Activate 激活员工账号
func (e *Employee) Activate() {
	e.IsActive = true
	e.UpdatedAt = time.Now()
}

// Deactivate 停用员工账号
func (e *Employee) Deactivate() {
	e.IsActive = false
	e.UpdatedAt = time.Now()
}

// TransferToMerchant 转移到其他商家
func (e *Employee) TransferToMerchant(newMerchantID int64) error {
	if newMerchantID <= 0 {
		return ErrInvalidMerchantID
	}

	e.MerchantID = newMerchantID
	e.UpdatedAt = time.Now()

	return nil
}

// #endregion

// #region 工具方法

// MaskEmail 脱敏显示邮箱
func (e *Employee) MaskEmail() string {
	return formatting.MaskEmail(e.Email)
}

// MaskPhone 脱敏显示手机号
func (e *Employee) MaskPhone() string {
	return formatting.MaskPhone(e.Phone)
}

// MaskIDNumber 脱敏显示身份证号
func (e *Employee) MaskIDNumber() string {
	if e.IDNumber == "" {
		return ""
	}

	if len(e.IDNumber) >= 10 {
		return e.IDNumber[:6] + "********" + e.IDNumber[len(e.IDNumber)-4:]
	}
	return e.IDNumber
}

// GetAge 根据身份证号计算年龄
func (e *Employee) GetAge() int {
	if e.IDNumber == "" || len(e.IDNumber) != 18 {
		return 0
	}

	birthYear := e.IDNumber[6:10]
	birthMonth := e.IDNumber[10:12]
	birthDay := e.IDNumber[12:14]

	// 精确的年龄计算
	birthDate := birthYear + "-" + birthMonth + "-" + birthDay
	if birth, err := time.Parse("2006-01-02", birthDate); err == nil {
		now := time.Now()
		age := now.Year() - birth.Year()
		if now.YearDay() < birth.YearDay() {
			age--
		}
		return age
	}

	return 0
}

// #endregion
