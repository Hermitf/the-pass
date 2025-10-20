package repository

import (
	"github.com/Hermitf/the-pass/internal/model"
	"gorm.io/gorm"
)

// #region 仓库定义

// RiderRepositoryInterface 配送员仓库接口
type RiderRepositoryInterface interface {
	// 基础CRUD
	Create(rider *model.Rider) error
	GetByID(id int64) (*model.Rider, error)
	Update(rider *model.Rider) error
	Delete(id int64) error

	// 查询方法
	GetByUsername(username string) (*model.Rider, error)
	GetByEmail(email string) (*model.Rider, error)
	GetByPhone(phone string) (*model.Rider, error)
	GetByIDNumber(idNumber string) (*model.Rider, error)
	GetByLicenseNumber(licenseNumber string) (*model.Rider, error)

	// 位置管理
	UpdateLocation(id int64, lat, lng float64) error
	GetRidersNearLocation(lat, lng, radiusKm float64) ([]*model.Rider, error)
	GetRidersByRegion(bounds map[string]float64) ([]*model.Rider, error)

	// 状态管理
	UpdateOnlineStatus(id int64, isOnline bool) error
	GetOnlineRiders(offset, limit int) ([]*model.Rider, int64, error)
	GetActiveRiders(offset, limit int) ([]*model.Rider, int64, error)
	GetAvailableRiders(lat, lng, radiusKm float64) ([]*model.Rider, error)

	// 列表查询
	GetRiderList(offset, limit int) ([]*model.Rider, int64, error)
	SearchRiders(keyword string, offset, limit int) ([]*model.Rider, int64, error)
	GetRidersByVehicleType(vehicleType string, offset, limit int) ([]*model.Rider, int64, error)

	// 业务查询方法
	CheckRiderExists(username, email, phone, licenseNumber string) (bool, error)
	GetRiderStats() (map[string]interface{}, error)
	GetTopRidersByRating(limit int) ([]*model.Rider, error)
	GetRidersByOrderCount(minOrders, maxOrders int64) ([]*model.Rider, error)
}

// RiderRepository 配送员仓库实现
type RiderRepository struct {
	db *gorm.DB
}

// NewRiderRepository 创建配送员仓库实例
func NewRiderRepository(db *gorm.DB) RiderRepositoryInterface {
	return &RiderRepository{
		db: db,
	}
}

// #endregion

// #region 基础CRUD操作

// Create 创建配送员
func (r *RiderRepository) Create(rider *model.Rider) error {
	if rider == nil {
		return ErrRiderNil
	}

	return r.db.Create(rider).Error
}

// GetByID 根据ID获取配送员
func (r *RiderRepository) GetByID(id int64) (*model.Rider, error) {
	if id <= 0 {
		return nil, ErrRiderIDInvalid
	}

	var rider model.Rider
	if err := r.db.Where("id = ?", id).First(&rider).Error; err != nil {
		return nil, err
	}
	return &rider, nil
}

// Update 更新配送员信息
func (r *RiderRepository) Update(rider *model.Rider) error {
	if rider == nil {
		return ErrRiderNil
	}

	return r.db.Save(rider).Error
}

// Delete 删除配送员（软删除）
func (r *RiderRepository) Delete(id int64) error {
	if id <= 0 {
		return ErrRiderIDInvalid
	}

	return r.db.Delete(&model.Rider{}, id).Error
}

// #endregion

// #region 查询方法

// GetByUsername 根据用户名获取配送员
func (r *RiderRepository) GetByUsername(username string) (*model.Rider, error) {
	if username == "" {
		return nil, ErrUsernameEmpty
	}

	var rider model.Rider
	if err := r.db.Where("username = ?", username).First(&rider).Error; err != nil {
		return nil, err
	}
	return &rider, nil
}

// GetByEmail 根据邮箱获取配送员
func (r *RiderRepository) GetByEmail(email string) (*model.Rider, error) {
	if email == "" {
		return nil, ErrEmailEmpty
	}

	var rider model.Rider
	if err := r.db.Where("email = ?", email).First(&rider).Error; err != nil {
		return nil, err
	}
	return &rider, nil
}

// GetByPhone 根据手机号获取配送员
func (r *RiderRepository) GetByPhone(phone string) (*model.Rider, error) {
	if phone == "" {
		return nil, ErrPhoneEmpty
	}

	var rider model.Rider
	if err := r.db.Where("phone = ?", phone).First(&rider).Error; err != nil {
		return nil, err
	}
	return &rider, nil
}

// GetByIDNumber 根据身份证号获取配送员
func (r *RiderRepository) GetByIDNumber(idNumber string) (*model.Rider, error) {
	if idNumber == "" {
		return nil, ErrIDNumberEmpty
	}

	var rider model.Rider
	if err := r.db.Where("id_number = ?", idNumber).First(&rider).Error; err != nil {
		return nil, err
	}
	return &rider, nil
}

// GetByLicenseNumber 根据驾照号获取配送员
func (r *RiderRepository) GetByLicenseNumber(licenseNumber string) (*model.Rider, error) {
	if licenseNumber == "" {
		return nil, ErrLicenseNumberEmpty
	}

	var rider model.Rider
	if err := r.db.Where("license_number = ?", licenseNumber).First(&rider).Error; err != nil {
		return nil, err
	}
	return &rider, nil
}

// #endregion

// #region 位置管理

// UpdateLocation 更新配送员位置
func (r *RiderRepository) UpdateLocation(id int64, lat, lng float64) error {
	if id <= 0 {
		return ErrRiderIDInvalid
	}

	// 验证位置范围
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return ErrLocationInvalid
	}

	return r.db.Model(&model.Rider{}).Where("id = ?", id).Updates(map[string]interface{}{
		"current_lat": lat,
		"current_lng": lng,
	}).Error
}

// GetRidersNearLocation 获取指定位置附近的配送员
func (r *RiderRepository) GetRidersNearLocation(lat, lng, radiusKm float64) ([]*model.Rider, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, ErrLocationInvalid
	}
	if radiusKm <= 0 {
		return nil, ErrRadiusInvalid
	}

	var riders []*model.Rider

	// 使用 Haversine 公式计算距离（MySQL版本）
	// 这里使用简化的矩形范围查询，实际项目中可以使用更精确的距离计算
	latRange := radiusKm / 111.0           // 大约每度纬度 111km
	lngRange := radiusKm / (111.0 * 0.707) // 经度范围（简化计算）

	if err := r.db.Where(
		"current_lat BETWEEN ? AND ? AND current_lng BETWEEN ? AND ? AND is_active = ? AND is_online = ?",
		lat-latRange, lat+latRange, lng-lngRange, lng+lngRange, true, true,
	).Find(&riders).Error; err != nil {
		return nil, err
	}

	return riders, nil
}

// GetRidersByRegion 根据地理边界获取配送员
func (r *RiderRepository) GetRidersByRegion(bounds map[string]float64) ([]*model.Rider, error) {
	minLat, hasMinLat := bounds["min_lat"]
	maxLat, hasMaxLat := bounds["max_lat"]
	minLng, hasMinLng := bounds["min_lng"]
	maxLng, hasMaxLng := bounds["max_lng"]

	if !hasMinLat || !hasMaxLat || !hasMinLng || !hasMaxLng {
		return nil, ErrBoundsInvalid
	}

	var riders []*model.Rider
	if err := r.db.Where(
		"current_lat BETWEEN ? AND ? AND current_lng BETWEEN ? AND ?",
		minLat, maxLat, minLng, maxLng,
	).Find(&riders).Error; err != nil {
		return nil, err
	}

	return riders, nil
}

// #endregion

// #region 状态管理

// UpdateOnlineStatus 更新在线状态
func (r *RiderRepository) UpdateOnlineStatus(id int64, isOnline bool) error {
	if id <= 0 {
		return ErrRiderIDInvalid
	}

	return r.db.Model(&model.Rider{}).Where("id = ?", id).Update("is_online", isOnline).Error
}

// GetOnlineRiders 获取在线配送员列表
func (r *RiderRepository) GetOnlineRiders(offset, limit int) ([]*model.Rider, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var riders []*model.Rider
	var total int64

	query := r.db.Model(&model.Rider{}).Where("is_online = ?", true)

	// 获取在线配送员总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询在线配送员
	if err := query.Offset(offset).Limit(limit).Find(&riders).Error; err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// GetActiveRiders 获取活跃配送员列表
func (r *RiderRepository) GetActiveRiders(offset, limit int) ([]*model.Rider, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var riders []*model.Rider
	var total int64

	query := r.db.Model(&model.Rider{}).Where("is_active = ?", true)

	// 获取活跃配送员总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询活跃配送员
	if err := query.Offset(offset).Limit(limit).Find(&riders).Error; err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// GetAvailableRiders 获取可接单的配送员（在线且活跃的附近配送员）
func (r *RiderRepository) GetAvailableRiders(lat, lng, radiusKm float64) ([]*model.Rider, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, ErrLocationInvalid
	}
	if radiusKm <= 0 {
		return nil, ErrRadiusInvalid
	}

	var riders []*model.Rider

	// 使用简化的矩形范围查询
	latRange := radiusKm / 111.0
	lngRange := radiusKm / (111.0 * 0.707)

	if err := r.db.Where(
		"current_lat BETWEEN ? AND ? AND current_lng BETWEEN ? AND ? AND is_active = ? AND is_online = ?",
		lat-latRange, lat+latRange, lng-lngRange, lng+lngRange, true, true,
	).Order("rating DESC").Find(&riders).Error; err != nil {
		return nil, err
	}

	return riders, nil
}

// #endregion

// #region 列表查询

// GetRiderList 分页获取配送员列表
func (r *RiderRepository) GetRiderList(offset, limit int) ([]*model.Rider, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var riders []*model.Rider
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Rider{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.Offset(offset).Limit(limit).Find(&riders).Error; err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// SearchRiders 搜索配送员
func (r *RiderRepository) SearchRiders(keyword string, offset, limit int) ([]*model.Rider, int64, error) {
	if keyword == "" {
		return r.GetRiderList(offset, limit)
	}

	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var riders []*model.Rider
	var total int64

	searchPattern := "%" + keyword + "%"
	query := r.db.Model(&model.Rider{}).Where(
		"username LIKE ? OR email LIKE ? OR phone LIKE ? OR name LIKE ? OR license_number LIKE ? OR vehicle_number LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
	)

	// 获取搜索结果总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询搜索结果
	if err := query.Offset(offset).Limit(limit).Find(&riders).Error; err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// GetRidersByVehicleType 根据交通工具类型获取配送员
func (r *RiderRepository) GetRidersByVehicleType(vehicleType string, offset, limit int) ([]*model.Rider, int64, error) {
	if vehicleType == "" {
		return nil, 0, ErrVehicleTypeEmpty
	}
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationParametersInvalid
	}

	var riders []*model.Rider
	var total int64

	query := r.db.Model(&model.Rider{}).Where("vehicle_type = ?", vehicleType)

	// 获取指定交通工具类型配送员总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询指定交通工具类型配送员
	if err := query.Offset(offset).Limit(limit).Find(&riders).Error; err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// #endregion

// #region 业务查询方法

// CheckRiderExists 检查配送员是否已存在
func (r *RiderRepository) CheckRiderExists(username, email, phone, licenseNumber string) (bool, error) {
	var count int64

	query := r.db.Model(&model.Rider{})
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
	if licenseNumber != "" {
		conditions = append(conditions, "license_number = ?")
		args = append(args, licenseNumber)
	}

	if len(conditions) == 0 {
		return false, ErrAtLeastOneFieldRequired
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

// GetRiderStats 获取配送员统计信息
func (r *RiderRepository) GetRiderStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总配送员数
	var totalRiders int64
	if err := r.db.Model(&model.Rider{}).Count(&totalRiders).Error; err != nil {
		return nil, err
	}
	stats["total_riders"] = totalRiders

	// 活跃配送员数
	var activeRiders int64
	if err := r.db.Model(&model.Rider{}).Where("is_active = ?", true).Count(&activeRiders).Error; err != nil {
		return nil, err
	}
	stats["active_riders"] = activeRiders

	// 在线配送员数
	var onlineRiders int64
	if err := r.db.Model(&model.Rider{}).Where("is_online = ?", true).Count(&onlineRiders).Error; err != nil {
		return nil, err
	}
	stats["online_riders"] = onlineRiders

	// 各交通工具类型统计
	var vehicleStats []map[string]interface{}
	if err := r.db.Model(&model.Rider{}).
		Select("vehicle_type, COUNT(*) as count").
		Group("vehicle_type").
		Scan(&vehicleStats).Error; err != nil {
		return nil, err
	}
	stats["vehicle_type_stats"] = vehicleStats

	// 今日新增配送员数
	var todayAdded int64
	if err := r.db.Model(&model.Rider{}).Where("DATE(created_at) = CURDATE()").Count(&todayAdded).Error; err != nil {
		return nil, err
	}
	stats["today_added"] = todayAdded

	return stats, nil
}

// GetTopRidersByRating 获取评分最高的配送员
func (r *RiderRepository) GetTopRidersByRating(limit int) ([]*model.Rider, error) {
	if limit <= 0 {
		return nil, ErrLimitInvalid
	}

	var riders []*model.Rider
	if err := r.db.Where("is_active = ?", true).
		Order("rating DESC, total_orders DESC").
		Limit(limit).Find(&riders).Error; err != nil {
		return nil, err
	}

	return riders, nil
}

// GetRidersByOrderCount 根据订单数量范围获取配送员
func (r *RiderRepository) GetRidersByOrderCount(minOrders, maxOrders int64) ([]*model.Rider, error) {
	if minOrders < 0 || maxOrders < minOrders {
		return nil, ErrOrderCountRangeInvalid
	}

	var riders []*model.Rider
	if err := r.db.Where("total_orders BETWEEN ? AND ?", minOrders, maxOrders).
		Order("total_orders DESC").Find(&riders).Error; err != nil {
		return nil, err
	}

	return riders, nil
}

// #endregion

// #region 工具方法

// GetRiderStatsByVehicleType 根据交通工具类型获取配送员统计
func (r *RiderRepository) GetRiderStatsByVehicleType() (map[string]int64, error) {
	stats := make(map[string]int64)

	var results []struct {
		VehicleType string `gorm:"column:vehicle_type"`
		Count       int64  `gorm:"column:count"`
	}

	if err := r.db.Model(&model.Rider{}).
		Select("vehicle_type, COUNT(*) as count").
		Group("vehicle_type").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, result := range results {
		stats[result.VehicleType] = result.Count
	}

	return stats, nil
}

// GetAverageRating 获取平均评分
func (r *RiderRepository) GetAverageRating() (float64, error) {
	var avgRating float64

	if err := r.db.Model(&model.Rider{}).
		Select("AVG(rating)").
		Where("total_orders > 0").
		Scan(&avgRating).Error; err != nil {
		return 0, err
	}

	return avgRating, nil
}

// GetMostExperiencedRiders 获取最有经验的配送员（订单数最多）
func (r *RiderRepository) GetMostExperiencedRiders(limit int) ([]*model.Rider, error) {
	if limit <= 0 {
		return nil, ErrLimitInvalid
	}

	var riders []*model.Rider
	if err := r.db.Where("is_active = ?", true).
		Order("total_orders DESC, rating DESC").
		Limit(limit).Find(&riders).Error; err != nil {
		return nil, err
	}

	return riders, nil
}

// #endregion
