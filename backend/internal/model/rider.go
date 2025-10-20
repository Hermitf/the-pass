package model

import (
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/Hermitf/the-pass/pkg/formatting"
	"gorm.io/gorm"
)

// #region 模型定义

// Rider 配送员模型
type Rider struct {
	ID            int64          `json:"id" gorm:"primaryKey;autoIncrement;comment:配送员ID"`
	Username      string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null;comment:用户名"`
	PasswordHash  string         `json:"-" gorm:"type:varchar(255);not null;comment:密码哈希"`
	Email         string         `json:"email" gorm:"type:varchar(100);uniqueIndex;not null;comment:邮箱"`
	Phone         string         `json:"phone" gorm:"type:varchar(20);uniqueIndex;not null;comment:手机号"`
	Name          string         `json:"name" gorm:"type:varchar(50);comment:真实姓名"`
	IDNumber      string         `json:"id_number" gorm:"type:varchar(20);uniqueIndex;comment:身份证号"`
	LicenseNumber string         `json:"license_number" gorm:"type:varchar(50);comment:驾照号"`
	VehicleType   string         `json:"vehicle_type" gorm:"type:varchar(20);comment:交通工具类型"` // bike, motorcycle, car
	VehicleNumber string         `json:"vehicle_number" gorm:"type:varchar(20);comment:车牌号"`
	CurrentLat    float64        `json:"current_lat" gorm:"comment:当前纬度"`
	CurrentLng    float64        `json:"current_lng" gorm:"comment:当前经度"`
	IsOnline      bool           `json:"is_online" gorm:"default:false;comment:是否在线"`
	IsActive      bool           `json:"is_active" gorm:"default:true;comment:是否激活"`
	Rating        float32        `json:"rating" gorm:"default:5.0;comment:评分"`
	TotalOrders   int64          `json:"total_orders" gorm:"default:0;comment:总订单数"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 设置表名
func (Rider) TableName() string {
	return "riders"
}

// #endregion

// #region 常量定义

// VehicleType 交通工具类型常量
const (
	VehicleTypeBike       = "bike"       // 自行车
	VehicleTypeMotorcycle = "motorcycle" // 摩托车
	VehicleTypeCar        = "car"        // 汽车
)

// 有效的交通工具类型列表
var ValidVehicleTypes = []string{
	VehicleTypeBike,
	VehicleTypeMotorcycle,
	VehicleTypeCar,
}

// #endregion

// #region 响应DTO

// RiderResponse 配送员响应DTO（用于API返回）
type RiderResponse struct {
	ID            int64   `json:"id"`
	Username      string  `json:"username"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	Name          string  `json:"name,omitempty"`
	VehicleType   string  `json:"vehicle_type,omitempty"`
	VehicleNumber string  `json:"vehicle_number,omitempty"`
	IsOnline      bool    `json:"is_online"`
	IsActive      bool    `json:"is_active"`
	Rating        float32 `json:"rating"`
	TotalOrders   int64   `json:"total_orders"`
}

// ToResponse 将 Rider 模型转换为响应DTO
func (r *Rider) ToResponse() *RiderResponse {
	return &RiderResponse{
		ID:            r.ID,
		Username:      r.Username,
		Email:         r.Email,
		Phone:         r.Phone,
		Name:          r.Name,
		VehicleType:   r.VehicleType,
		VehicleNumber: r.VehicleNumber,
		IsOnline:      r.IsOnline,
		IsActive:      r.IsActive,
		Rating:        r.Rating,
		TotalOrders:   r.TotalOrders,
	}
}

// RiderLocationResponse 配送员位置响应DTO
type RiderLocationResponse struct {
	ID         int64   `json:"id"`
	Username   string  `json:"username"`
	Name       string  `json:"name,omitempty"`
	CurrentLat float64 `json:"current_lat"`
	CurrentLng float64 `json:"current_lng"`
	IsOnline   bool    `json:"is_online"`
	Rating     float32 `json:"rating"`
}

// ToLocationResponse 转换为位置响应DTO
func (r *Rider) ToLocationResponse() *RiderLocationResponse {
	return &RiderLocationResponse{
		ID:         r.ID,
		Username:   r.Username,
		Name:       r.Name,
		CurrentLat: r.CurrentLat,
		CurrentLng: r.CurrentLng,
		IsOnline:   r.IsOnline,
		Rating:     r.Rating,
	}
}

// RiderSafeResponse 配送员安全响应DTO（脱敏显示）
type RiderSafeResponse struct {
	ID            int64   `json:"id"`
	Username      string  `json:"username"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	Name          string  `json:"name,omitempty"`
	VehicleType   string  `json:"vehicle_type,omitempty"`
	VehicleNumber string  `json:"vehicle_number,omitempty"` // 脱敏显示
	IsOnline      bool    `json:"is_online"`
	IsActive      bool    `json:"is_active"`
	Rating        float32 `json:"rating"`
	TotalOrders   int64   `json:"total_orders"`
}

// ToSafeResponse 转换为安全响应（脱敏）
func (r *Rider) ToSafeResponse() *RiderSafeResponse {
	return &RiderSafeResponse{
		ID:            r.ID,
		Username:      r.Username,
		Email:         r.MaskEmail(),
		Phone:         r.MaskPhone(),
		Name:          r.Name,
		VehicleType:   r.VehicleType,
		VehicleNumber: r.MaskVehicleNumber(),
		IsOnline:      r.IsOnline,
		IsActive:      r.IsActive,
		Rating:        r.Rating,
		TotalOrders:   r.TotalOrders,
	}
}

// #endregion

// #region 验证方法

// ValidateVehicleType 验证交通工具类型
func (r *Rider) ValidateVehicleType() error {
	if r.VehicleType == "" {
		return nil // 交通工具类型可选
	}

	for _, validType := range ValidVehicleTypes {
		if r.VehicleType == validType {
			return nil
		}
	}
	return ErrInvalidVehicleType
}

// ValidateVehicleNumber 验证车牌号
func (r *Rider) ValidateVehicleNumber() error {
	if r.VehicleNumber == "" {
		return nil // 车牌号可选
	}

	// 简单的车牌号验证（中国车牌格式）
	matched, _ := regexp.MatchString(`^[京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领A-Z]{1}[A-Z]{1}[A-Z0-9]{4}[A-Z0-9挂学警港澳]{1}$`, r.VehicleNumber)
	if !matched {
		return ErrInvalidVehicleNumber
	}
	return nil
}

// ValidateLicenseNumber 验证驾照号
func (r *Rider) ValidateLicenseNumber() error {
	if r.LicenseNumber == "" {
		return nil // 驾照号可选
	}

	// 驾照号验证（通常是18位数字）
	matched, _ := regexp.MatchString(`^\d{18}$`, r.LicenseNumber)
	if !matched {
		return ErrInvalidLicenseNumber
	}
	return nil
}

// ValidateLocation 验证位置信息
func (r *Rider) ValidateLocation() error {
	// 纬度范围 -90 到 90
	if r.CurrentLat < -90 || r.CurrentLat > 90 {
		return ErrInvalidLocation
	}

	// 经度范围 -180 到 180
	if r.CurrentLng < -180 || r.CurrentLng > 180 {
		return ErrInvalidLocation
	}

	return nil
}

// ValidateAll 验证所有字段
func (r *Rider) ValidateAll() error {
	if err := r.ValidateVehicleType(); err != nil {
		return err
	}
	if err := r.ValidateVehicleNumber(); err != nil {
		return err
	}
	if err := r.ValidateLicenseNumber(); err != nil {
		return err
	}
	if err := r.ValidateLocation(); err != nil {
		return err
	}
	return nil
}

// #endregion

// #region 位置管理

// UpdateLocation 更新配送员位置
func (r *Rider) UpdateLocation(lat, lng float64) error {
	// 创建临时对象验证位置
	tempRider := &Rider{
		CurrentLat: lat,
		CurrentLng: lng,
	}

	if err := tempRider.ValidateLocation(); err != nil {
		return err
	}

	r.CurrentLat = lat
	r.CurrentLng = lng
	r.UpdatedAt = time.Now()

	return nil
}

// GetCurrentLocation 获取当前位置
func (r *Rider) GetCurrentLocation() (float64, float64) {
	return r.CurrentLat, r.CurrentLng
}

// CalculateDistance 计算与指定位置的距离（公里）
func (r *Rider) CalculateDistance(lat, lng float64) float64 {
	// 使用 Haversine 公式计算两点间距离
	const earthRadius = 6371.0 // 地球半径（公里）

	lat1 := r.CurrentLat * math.Pi / 180
	lng1 := r.CurrentLng * math.Pi / 180
	lat2 := lat * math.Pi / 180
	lng2 := lng * math.Pi / 180

	dlat := lat2 - lat1
	dlng := lng2 - lng1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// IsNearLocation 检查是否在指定位置附近（范围内，单位：公里）
func (r *Rider) IsNearLocation(lat, lng, radiusKm float64) bool {
	distance := r.CalculateDistance(lat, lng)
	return distance <= radiusKm
}

// #endregion

// #region 状态管理

// GoOnline 配送员上线
func (r *Rider) GoOnline() error {
	if !r.IsActive {
		return ErrRiderNotActive
	}

	r.IsOnline = true
	r.UpdatedAt = time.Now()
	return nil
}

// GoOffline 配送员下线
func (r *Rider) GoOffline() {
	r.IsOnline = false
	r.UpdatedAt = time.Now()
}

// IsAvailableForOrder 检查是否可接订单
func (r *Rider) IsAvailableForOrder() bool {
	return r.IsActive && r.IsOnline
}

// Activate 激活配送员账号
func (r *Rider) Activate() {
	r.IsActive = true
	r.UpdatedAt = time.Now()
}

// Deactivate 停用配送员账号
func (r *Rider) Deactivate() {
	r.IsActive = false
	r.IsOnline = false // 停用时同时下线
	r.UpdatedAt = time.Now()
}

// #endregion

// #region 业务方法

// GetDisplayName 获取显示名称
func (r *Rider) GetDisplayName() string {
	if r.Name != "" {
		return r.Name
	}
	return r.Username
}

// UpdateProfile 更新配送员基本信息
func (r *Rider) UpdateProfile(name, vehicleType, vehicleNumber, licenseNumber string) error {
	// 创建临时对象进行验证
	tempRider := &Rider{
		Name:          name,
		VehicleType:   vehicleType,
		VehicleNumber: vehicleNumber,
		LicenseNumber: licenseNumber,
	}

	if err := tempRider.ValidateAll(); err != nil {
		return err
	}

	r.Name = name
	r.VehicleType = vehicleType
	r.VehicleNumber = vehicleNumber
	r.LicenseNumber = licenseNumber
	r.UpdatedAt = time.Now()

	return nil
}

// CompleteOrder 完成订单（更新统计信息）
func (r *Rider) CompleteOrder(rating float32) {
	r.TotalOrders++

	// 更新评分（简单的加权平均）
	if r.TotalOrders == 1 {
		r.Rating = rating
	} else {
		// 使用加权平均计算新评分
		oldWeight := float32(r.TotalOrders - 1)
		r.Rating = (r.Rating*oldWeight + rating) / float32(r.TotalOrders)
	}

	r.UpdatedAt = time.Now()
}

// GetVehicleTypeDisplay 获取交通工具类型的显示名称
func (r *Rider) GetVehicleTypeDisplay() string {
	switch r.VehicleType {
	case VehicleTypeBike:
		return "自行车"
	case VehicleTypeMotorcycle:
		return "摩托车"
	case VehicleTypeCar:
		return "汽车"
	default:
		return r.VehicleType
	}
}

// #endregion

// #region 工具方法

// MaskEmail 脱敏显示邮箱
func (r *Rider) MaskEmail() string {
	return formatting.MaskEmail(r.Email)
}

// MaskPhone 脱敏显示手机号
func (r *Rider) MaskPhone() string {
	return formatting.MaskPhone(r.Phone)
}

// MaskVehicleNumber 脱敏显示车牌号
func (r *Rider) MaskVehicleNumber() string {
	if r.VehicleNumber == "" {
		return ""
	}

	if len(r.VehicleNumber) >= 4 {
		return r.VehicleNumber[:2] + "***" + r.VehicleNumber[len(r.VehicleNumber)-1:]
	}
	return r.VehicleNumber
}

// GetRiderStats 获取配送员统计信息
func (r *Rider) GetRiderStats() map[string]interface{} {
	return map[string]interface{}{
		"id":               r.ID,
		"username":         r.Username,
		"name":             r.GetDisplayName(),
		"vehicle_type":     r.GetVehicleTypeDisplay(),
		"is_online":        r.IsOnline,
		"is_active":        r.IsActive,
		"rating":           r.Rating,
		"total_orders":     r.TotalOrders,
		"current_location": fmt.Sprintf("(%.6f, %.6f)", r.CurrentLat, r.CurrentLng),
		"created_at":       r.CreatedAt,
	}
}

// #endregion
