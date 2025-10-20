package repository

import (
	"fmt"

	"github.com/Hermitf/the-pass/internal/model"
	"gorm.io/gorm"
)

// #region 仓库定义

// EmployeeRepositoryInterface 员工仓库接口
type EmployeeRepositoryInterface interface {
	// 基础CRUD操作
	Create(employee *model.Employee) error
	GetByID(id int64) (*model.Employee, error)
	Update(employee *model.Employee) error
	Delete(id int64) error

	// 查询方法
	GetByUsername(username string) (*model.Employee, error)
	GetByEmail(email string) (*model.Employee, error)
	GetByPhone(phone string) (*model.Employee, error)
	GetByIDNumber(idNumber string) (*model.Employee, error)

	// 商家关联查询
	GetByMerchantID(merchantID int64) ([]*model.Employee, error)
	GetActiveEmployeesByMerchant(merchantID int64) ([]*model.Employee, error)
	GetEmployeesByMerchantWithPagination(merchantID int64, offset, limit int) ([]*model.Employee, int64, error)

	// 业务查询方法
	CheckEmployeeExists(username, email, phone string) (bool, error)
	SearchEmployees(keyword string, merchantID int64, offset, limit int) ([]*model.Employee, int64, error)
	GetEmployeeStatsByMerchant(merchantID int64) (map[string]interface{}, error)

	// 员工转移
	TransferEmployee(employeeID, newMerchantID int64) error
	BulkTransferEmployees(employeeIDs []int64, newMerchantID int64) error
}

// EmployeeRepository 员工仓库实现
type EmployeeRepository struct {
	db *gorm.DB
}

// #endregion

// #region 构造函数

// NewEmployeeRepository 创建员工仓库实例
func NewEmployeeRepository(db *gorm.DB) EmployeeRepositoryInterface {
	return &EmployeeRepository{
		db: db,
	}
}

// #endregion

// #region 基础CRUD操作

// Create 创建员工
func (r *EmployeeRepository) Create(employee *model.Employee) error {
	if employee == nil {
		return ErrEmployeeNil
	}

	return r.db.Create(employee).Error
}

// GetByID 根据ID获取员工
func (r *EmployeeRepository) GetByID(id int64) (*model.Employee, error) {
	if id <= 0 {
		return nil, ErrEmployeeIDInvalid
	}

	var employee model.Employee
	if err := r.db.Where("id = ?", id).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// Update 更新员工信息
func (r *EmployeeRepository) Update(employee *model.Employee) error {
	if employee == nil {
		return ErrEmployeeNil
	}

	return r.db.Save(employee).Error
}

// Delete 删除员工（软删除）
func (r *EmployeeRepository) Delete(id int64) error {
	if id <= 0 {
		return ErrEmployeeIDInvalid
	}

	return r.db.Delete(&model.Employee{}, id).Error
}

// #endregion

// #region 查询方法

// GetByUsername 根据用户名获取员工
func (r *EmployeeRepository) GetByUsername(username string) (*model.Employee, error) {
	if username == "" {
		return nil, ErrUsernameEmpty
	}

	var employee model.Employee
	if err := r.db.Where("username = ?", username).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetByEmail 根据邮箱获取员工
func (r *EmployeeRepository) GetByEmail(email string) (*model.Employee, error) {
	if email == "" {
		return nil, ErrEmailEmpty
	}

	var employee model.Employee
	if err := r.db.Where("email = ?", email).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetByPhone 根据手机号获取员工
func (r *EmployeeRepository) GetByPhone(phone string) (*model.Employee, error) {
	if phone == "" {
		return nil, ErrPhoneEmpty
	}

	var employee model.Employee
	if err := r.db.Where("phone = ?", phone).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetByIDNumber 根据身份证号获取员工
func (r *EmployeeRepository) GetByIDNumber(idNumber string) (*model.Employee, error) {
	if idNumber == "" {
		return nil, ErrIDNumberEmpty
	}

	var employee model.Employee
	if err := r.db.Where("id_number = ?", idNumber).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// #endregion

// #region 商家关联查询

// GetByMerchantID 根据商家ID获取员工列表
func (r *EmployeeRepository) GetByMerchantID(merchantID int64) ([]*model.Employee, error) {
	if merchantID <= 0 {
		return nil, ErrMerchantIDInvalid
	}

	var employees []*model.Employee
	if err := r.db.Where("merchant_id = ?", merchantID).Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

// GetActiveEmployeesByMerchant 获取商家的活跃员工列表
func (r *EmployeeRepository) GetActiveEmployeesByMerchant(merchantID int64) ([]*model.Employee, error) {
	if merchantID <= 0 {
		return nil, ErrMerchantIDInvalid
	}

	var employees []*model.Employee
	if err := r.db.Where("merchant_id = ? AND is_active = ?", merchantID, true).Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

// GetEmployeesByMerchantWithPagination 分页获取商家员工列表
func (r *EmployeeRepository) GetEmployeesByMerchantWithPagination(merchantID int64, offset, limit int) ([]*model.Employee, int64, error) {
	if merchantID <= 0 {
		return nil, 0, ErrMerchantIDInvalid
	}
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	var employees []*model.Employee
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Employee{}).Where("merchant_id = ?", merchantID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.Where("merchant_id = ?", merchantID).Offset(offset).Limit(limit).Find(&employees).Error; err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// #endregion

// #region 业务查询方法

// CheckEmployeeExists 检查员工是否已存在
func (r *EmployeeRepository) CheckEmployeeExists(username, email, phone string) (bool, error) {
	var count int64

	query := r.db.Model(&model.Employee{})
	conditions := []string{}
	args := []interface{}{}

	if username != "" {
		conditions = append(conditions, "username = ?")
		args = append(args, username)
	}
	if email != "" {
		conditions = append(conditions, "email = ?")
		args = append(args, email)
	}
	if phone != "" {
		conditions = append(conditions, "phone = ?")
		args = append(args, phone)
	}

	if len(conditions) == 0 {
		return false, ErrEmployeeUpdateFieldsEmpty
	}

	whereClause := conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " OR " + conditions[i]
	}

	if err := query.Where(whereClause, args...).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// SearchEmployees 搜索员工
func (r *EmployeeRepository) SearchEmployees(keyword string, merchantID int64, offset, limit int) ([]*model.Employee, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var employees []*model.Employee
	var total int64

	query := r.db.Model(&model.Employee{})

	// 添加商家ID过滤条件
	if merchantID > 0 {
		query = query.Where("merchant_id = ?", merchantID)
	}

	// 添加关键字搜索条件
	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where(
			"username LIKE ? OR email LIKE ? OR phone LIKE ? OR name LIKE ? OR id_number LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// 获取搜索结果总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询搜索结果
	if err := query.Offset(offset).Limit(limit).Find(&employees).Error; err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// GetEmployeeStatsByMerchant 获取商家员工统计信息
func (r *EmployeeRepository) GetEmployeeStatsByMerchant(merchantID int64) (map[string]interface{}, error) {
	if merchantID <= 0 {
		return nil, ErrMerchantIDInvalid
	}

	stats := make(map[string]interface{})

	// 总员工数
	var totalEmployees int64
	if err := r.db.Model(&model.Employee{}).Where("merchant_id = ?", merchantID).Count(&totalEmployees).Error; err != nil {
		return nil, err
	}
	stats["total_employees"] = totalEmployees

	// 活跃员工数
	var activeEmployees int64
	if err := r.db.Model(&model.Employee{}).Where("merchant_id = ? AND is_active = ?", merchantID, true).Count(&activeEmployees).Error; err != nil {
		return nil, err
	}
	stats["active_employees"] = activeEmployees

	// 今日新增员工数
	var todayAdded int64
	if err := r.db.Model(&model.Employee{}).Where("merchant_id = ? AND DATE(created_at) = CURDATE()", merchantID).Count(&todayAdded).Error; err != nil {
		return nil, err
	}
	stats["today_added"] = todayAdded

	return stats, nil
}

// #endregion

// #region 员工转移

// TransferEmployee 转移单个员工到新商家
func (r *EmployeeRepository) TransferEmployee(employeeID, newMerchantID int64) error {
	if employeeID <= 0 {
		return ErrEmployeeIDInvalid
	}
	if newMerchantID <= 0 {
		return ErrMerchantIDInvalid
	}

	return r.db.Model(&model.Employee{}).Where("id = ?", employeeID).Update("merchant_id", newMerchantID).Error
}

// BulkTransferEmployees 批量转移员工到新商家
func (r *EmployeeRepository) BulkTransferEmployees(employeeIDs []int64, newMerchantID int64) error {
	if len(employeeIDs) == 0 {
		return ErrEmployeeIDsEmpty
	}
	if newMerchantID <= 0 {
		return ErrMerchantIDInvalid
	}

	return r.db.Model(&model.Employee{}).Where("id IN ?", employeeIDs).Update("merchant_id", newMerchantID).Error
}

// #endregion

// #region 工具方法

// CountEmployeesByMerchant 统计商家员工数量
func (r *EmployeeRepository) CountEmployeesByMerchant(merchantID int64) (int64, error) {
	if merchantID <= 0 {
		return 0, ErrMerchantIDInvalid
	}

	var count int64
	if err := r.db.Model(&model.Employee{}).Where("merchant_id = ?", merchantID).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// GetEmployeesByAge 根据年龄范围获取员工
func (r *EmployeeRepository) GetEmployeesByAge(merchantID int64, minAge, maxAge int) ([]*model.Employee, error) {
	if merchantID <= 0 {
		return nil, ErrMerchantIDInvalid
	}
	if minAge < 0 || maxAge < minAge {
		return nil, ErrAgeRangeInvalid
	}

	var employees []*model.Employee
	// 注意：这里使用简单的年份计算，实际应用中可能需要更精确的年龄计算
	currentYear := "YEAR(CURDATE())"
	birthYearFromID := "CAST(SUBSTR(id_number, 7, 4) AS UNSIGNED)"

	query := fmt.Sprintf(
		"merchant_id = ? AND (%s - %s) BETWEEN ? AND ?",
		currentYear, birthYearFromID,
	)

	if err := r.db.Where(query, merchantID, minAge, maxAge).Find(&employees).Error; err != nil {
		return nil, err
	}

	return employees, nil
}

// GetRecentlyJoinedEmployees 获取最近加入的员工
func (r *EmployeeRepository) GetRecentlyJoinedEmployees(merchantID int64, days int) ([]*model.Employee, error) {
	if merchantID <= 0 {
		return nil, ErrMerchantIDInvalid
	}
	if days <= 0 {
		return nil, ErrDaysInvalid
	}

	var employees []*model.Employee
	if err := r.db.Where(
		"merchant_id = ? AND created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)",
		merchantID, days,
	).Order("created_at DESC").Find(&employees).Error; err != nil {
		return nil, err
	}

	return employees, nil
}

// #endregion
