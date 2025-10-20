package repository

import (
	"github.com/Hermitf/the-pass/internal/model"
	"gorm.io/gorm"
)

// #region 仓库定义

// MerchantRepositoryInterface 商家仓库接口
type MerchantRepositoryInterface interface {
	// 基础CRUD操作
	Create(merchant *model.Merchant) error
	GetByID(id int64) (*model.Merchant, error)
	Update(merchant *model.Merchant) error
	Delete(id int64) error

	// 查询方法
	GetByUsername(username string) (*model.Merchant, error)
	GetByEmail(email string) (*model.Merchant, error)
	GetByPhone(phone string) (*model.Merchant, error)
	GetByBusinessLicense(license string) (*model.Merchant, error)

	// 列表查询
	GetMerchantList(offset, limit int) ([]*model.Merchant, int64, error)
	GetActiveMerchants(offset, limit int) ([]*model.Merchant, int64, error)
	SearchMerchants(keyword string, offset, limit int) ([]*model.Merchant, int64, error)

	// 员工关联查询
	GetMerchantWithEmployees(id int64) (*model.Merchant, []*model.Employee, error)
	GetMerchantsByEmployeeCount(minCount, maxCount int) ([]*model.Merchant, error)

	// 业务查询方法
	CheckMerchantExists(username, email, phone, businessLicense string) (bool, error)
	GetMerchantStats() (map[string]interface{}, error)
	GetMerchantsByRegion(region string, offset, limit int) ([]*model.Merchant, int64, error)
	GetTopMerchantsByEmployeeCount(limit int) ([]*model.Merchant, error)
}

// MerchantRepository 商家仓库实现
type MerchantRepository struct {
	db *gorm.DB
}

// #endregion

// #region 构造函数

// NewMerchantRepository 创建商家仓库实例
func NewMerchantRepository(db *gorm.DB) MerchantRepositoryInterface {
	return &MerchantRepository{
		db: db,
	}
}

// #endregion

// #region 基础CRUD操作

// Create 创建商家
func (r *MerchantRepository) Create(merchant *model.Merchant) error {
	if merchant == nil {
		return ErrMerchantNil
	}

	return r.db.Create(merchant).Error
}

// GetByID 根据ID获取商家
func (r *MerchantRepository) GetByID(id int64) (*model.Merchant, error) {
	if id <= 0 {
		return nil, ErrMerchantIDInvalid
	}

	var merchant model.Merchant
	if err := r.db.Where("id = ?", id).First(&merchant).Error; err != nil {
		return nil, err
	}
	return &merchant, nil
}

// Update 更新商家信息
func (r *MerchantRepository) Update(merchant *model.Merchant) error {
	if merchant == nil {
		return ErrMerchantNil
	}

	return r.db.Save(merchant).Error
}

// Delete 删除商家（软删除）
func (r *MerchantRepository) Delete(id int64) error {
	if id <= 0 {
		return ErrMerchantIDInvalid
	}

	return r.db.Delete(&model.Merchant{}, id).Error
}

// #endregion

// #region 查询方法

// GetByUsername 根据用户名获取商家
func (r *MerchantRepository) GetByUsername(username string) (*model.Merchant, error) {
	if username == "" {
		return nil, ErrUsernameEmpty
	}

	var merchant model.Merchant
	if err := r.db.Where("username = ?", username).First(&merchant).Error; err != nil {
		return nil, err
	}
	return &merchant, nil
}

// GetByEmail 根据邮箱获取商家
func (r *MerchantRepository) GetByEmail(email string) (*model.Merchant, error) {
	if email == "" {
		return nil, ErrEmailEmpty
	}

	var merchant model.Merchant
	if err := r.db.Where("email = ?", email).First(&merchant).Error; err != nil {
		return nil, err
	}
	return &merchant, nil
}

// GetByPhone 根据手机号获取商家
func (r *MerchantRepository) GetByPhone(phone string) (*model.Merchant, error) {
	if phone == "" {
		return nil, ErrPhoneEmpty
	}

	var merchant model.Merchant
	if err := r.db.Where("phone = ?", phone).First(&merchant).Error; err != nil {
		return nil, err
	}
	return &merchant, nil
}

// GetByBusinessLicense 根据营业执照号获取商家
func (r *MerchantRepository) GetByBusinessLicense(license string) (*model.Merchant, error) {
	if license == "" {
		return nil, ErrBusinessLicenseEmpty
	}

	var merchant model.Merchant
	if err := r.db.Where("business_license = ?", license).First(&merchant).Error; err != nil {
		return nil, err
	}
	return &merchant, nil
}

// #endregion

// #region 列表查询

// GetMerchantList 分页获取商家列表
func (r *MerchantRepository) GetMerchantList(offset, limit int) ([]*model.Merchant, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	var merchants []*model.Merchant
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Merchant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.Offset(offset).Limit(limit).Find(&merchants).Error; err != nil {
		return nil, 0, err
	}

	return merchants, total, nil
}

// GetActiveMerchants 获取活跃商家列表
func (r *MerchantRepository) GetActiveMerchants(offset, limit int) ([]*model.Merchant, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var merchants []*model.Merchant
	var total int64

	// 获取活跃商家总数
	if err := r.db.Model(&model.Merchant{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询活跃商家
	if err := r.db.Where("is_active = ?", true).Offset(offset).Limit(limit).Find(&merchants).Error; err != nil {
		return nil, 0, err
	}

	return merchants, total, nil
}

// SearchMerchants 搜索商家
func (r *MerchantRepository) SearchMerchants(keyword string, offset, limit int) ([]*model.Merchant, int64, error) {
	if keyword == "" {
		return r.GetMerchantList(offset, limit)
	}

	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var merchants []*model.Merchant
	var total int64

	searchPattern := "%" + keyword + "%"
	query := r.db.Model(&model.Merchant{}).Where(
		"username LIKE ? OR email LIKE ? OR phone LIKE ? OR company_name LIKE ? OR business_license LIKE ? OR address LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
	)

	// 获取搜索结果总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询搜索结果
	if err := query.Offset(offset).Limit(limit).Find(&merchants).Error; err != nil {
		return nil, 0, err
	}

	return merchants, total, nil
}

// #endregion

// #region 员工关联查询

// GetMerchantWithEmployees 获取商家及其员工信息
func (r *MerchantRepository) GetMerchantWithEmployees(id int64) (*model.Merchant, []*model.Employee, error) {
	if id <= 0 {
		return nil, nil, ErrMerchantIDInvalid
	}

	var merchant model.Merchant
	if err := r.db.Where("id = ?", id).First(&merchant).Error; err != nil {
		return nil, nil, err
	}

	var employees []*model.Employee
	if err := r.db.Where("merchant_id = ?", id).Find(&employees).Error; err != nil {
		return &merchant, nil, err
	}

	return &merchant, employees, nil
}

// GetMerchantsByEmployeeCount 根据员工数量范围获取商家
func (r *MerchantRepository) GetMerchantsByEmployeeCount(minCount, maxCount int) ([]*model.Merchant, error) {
	if minCount < 0 || maxCount < minCount {
		return nil, ErrEmployeeCountInvalid
	}

	var merchants []*model.Merchant

	subQuery := r.db.Model(&model.Employee{}).
		Select("merchant_id, COUNT(*) as employee_count").
		Group("merchant_id").
		Having("employee_count BETWEEN ? AND ?", minCount, maxCount)

	if err := r.db.Where("id IN (?)", subQuery).Find(&merchants).Error; err != nil {
		return nil, err
	}

	return merchants, nil
}

// #endregion

// #region 业务查询方法

// CheckMerchantExists 检查商家是否已存在
func (r *MerchantRepository) CheckMerchantExists(username, email, phone, businessLicense string) (bool, error) {
	var count int64

	query := r.db.Model(&model.Merchant{})
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
	if businessLicense != "" {
		conditions = append(conditions, "business_license = ?")
		args = append(args, businessLicense)
	}

	if len(conditions) == 0 {
		return false, ErrAtLeastOneField
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

// GetMerchantStats 获取商家统计信息
func (r *MerchantRepository) GetMerchantStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总商家数
	var totalMerchants int64
	if err := r.db.Model(&model.Merchant{}).Count(&totalMerchants).Error; err != nil {
		return nil, err
	}
	stats["total_merchants"] = totalMerchants

	// 活跃商家数
	var activeMerchants int64
	if err := r.db.Model(&model.Merchant{}).Where("is_active = ?", true).Count(&activeMerchants).Error; err != nil {
		return nil, err
	}
	stats["active_merchants"] = activeMerchants

	// 今日注册商家数
	var todayRegistered int64
	if err := r.db.Model(&model.Merchant{}).Where("DATE(created_at) = CURDATE()").Count(&todayRegistered).Error; err != nil {
		return nil, err
	}
	stats["today_registered"] = todayRegistered

	// 本月注册商家数
	var monthRegistered int64
	if err := r.db.Model(&model.Merchant{}).Where("YEAR(created_at) = YEAR(CURDATE()) AND MONTH(created_at) = MONTH(CURDATE())").Count(&monthRegistered).Error; err != nil {
		return nil, err
	}
	stats["month_registered"] = monthRegistered

	return stats, nil
}

// GetMerchantsByRegion 根据地区获取商家
func (r *MerchantRepository) GetMerchantsByRegion(region string, offset, limit int) ([]*model.Merchant, int64, error) {
	if region == "" {
		return nil, 0, ErrRegionEmpty
	}
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var merchants []*model.Merchant
	var total int64

	regionPattern := "%" + region + "%"
	query := r.db.Model(&model.Merchant{}).Where("address LIKE ?", regionPattern)

	// 获取指定地区商家总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询指定地区商家
	if err := query.Offset(offset).Limit(limit).Find(&merchants).Error; err != nil {
		return nil, 0, err
	}

	return merchants, total, nil
}

// #endregion

// #region 工具方法

// IsUsernameAvailable 检查用户名是否可用
func (r *MerchantRepository) IsUsernameAvailable(username string) (bool, error) {
	if username == "" {
		return false, ErrUsernameEmpty
	}

	var count int64
	if err := r.db.Model(&model.Merchant{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}

	return count == 0, nil
}

// IsEmailAvailable 检查邮箱是否可用
func (r *MerchantRepository) IsEmailAvailable(email string) (bool, error) {
	if email == "" {
		return false, ErrEmailEmpty
	}

	var count int64
	if err := r.db.Model(&model.Merchant{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}

	return count == 0, nil
}

// IsPhoneAvailable 检查手机号是否可用
func (r *MerchantRepository) IsPhoneAvailable(phone string) (bool, error) {
	if phone == "" {
		return false, ErrPhoneEmpty
	}

	var count int64
	if err := r.db.Model(&model.Merchant{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return false, err
	}

	return count == 0, nil
}

// IsBusinessLicenseAvailable 检查营业执照号是否可用
func (r *MerchantRepository) IsBusinessLicenseAvailable(license string) (bool, error) {
	if license == "" {
		return false, ErrBusinessLicenseEmpty
	}

	var count int64
	if err := r.db.Model(&model.Merchant{}).Where("business_license = ?", license).Count(&count).Error; err != nil {
		return false, err
	}

	return count == 0, nil
}

// GetTopMerchantsByEmployeeCount 获取员工数量最多的商家（排行榜）
func (r *MerchantRepository) GetTopMerchantsByEmployeeCount(limit int) ([]*model.Merchant, error) {
	if limit <= 0 {
		return nil, ErrLimitInvalid
	}

	var merchants []*model.Merchant

	subQuery := r.db.Model(&model.Employee{}).
		Select("merchant_id, COUNT(*) as employee_count").
		Group("merchant_id").
		Order("employee_count DESC").
		Limit(limit)

	if err := r.db.Where("id IN (?)", subQuery).Find(&merchants).Error; err != nil {
		return nil, err
	}

	return merchants, nil
}

// #endregion
