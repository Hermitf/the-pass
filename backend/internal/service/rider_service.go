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

// RiderServiceInterface 配送员服务接口
type RiderServiceInterface interface {
	// 配送员注册和认证
	RegisterRider(rider *model.Rider) error
	LoginRider(loginInfo, password, loginType string) (string, error)

	// 配送员信息管理
	GetRiderByID(id int64) (*model.Rider, error)
	UpdateRiderProfile(riderID int64, name, vehicleType, vehicleNumber, licenseNumber string) error
	UpdateRiderPassword(riderID int64, oldPassword, newPassword string) error

	// 位置管理
	UpdateLocation(riderID int64, lat, lng float64) error
	GetRidersNearLocation(lat, lng, radiusKm float64) ([]*model.Rider, error)
	GetRidersByRegion(bounds map[string]float64) ([]*model.Rider, error)

	// 状态管理
	SetOnlineStatus(riderID int64, isOnline bool) error
	GetOnlineRiders(offset, limit int) ([]*model.Rider, int64, error)
	GetActiveRiders(offset, limit int) ([]*model.Rider, int64, error)
	GetAvailableRiders(lat, lng, radiusKm float64) ([]*model.Rider, error)

	// 配送员验证
	ValidateRiderData(rider *model.Rider) error
	CheckRiderAvailability(username, email, phone, licenseNumber string) error

	// 配送员列表和搜索
	GetRiderList(offset, limit int) ([]*model.Rider, int64, error)
	SearchRiders(keyword string, offset, limit int) ([]*model.Rider, int64, error)
	GetRidersByVehicleType(vehicleType string, offset, limit int) ([]*model.Rider, int64, error)

	// 配送员统计
	GetRiderStats() (map[string]interface{}, error)
	GetTopRidersByRating(limit int) ([]*model.Rider, error)
	GetRidersByOrderCount(minOrders, maxOrders int64) ([]*model.Rider, error)
}

// RiderService 配送员服务实现
type RiderService struct {
	riderRepo  repository.RiderRepositoryInterface
	jwtService JWTServiceInterface
}

// #endregion

// #region 构造函数和依赖注入

// RiderServiceDependencies 配送员服务依赖
type RiderServiceDependencies struct {
	RiderRepo  repository.RiderRepositoryInterface
	JWTService JWTServiceInterface
}

// NewRiderService 创建配送员服务实例
func NewRiderService(deps RiderServiceDependencies) RiderServiceInterface {
	return &RiderService{
		riderRepo:  deps.RiderRepo,
		jwtService: deps.JWTService,
	}
}

// #endregion

// #region 配送员注册和认证

// RegisterRider 注册配送员
func (s *RiderService) RegisterRider(rider *model.Rider) error {
	if rider == nil {
		return ErrRiderNil
	}

	// 验证配送员数据
	if err := s.ValidateRiderData(rider); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 检查配送员是否已存在
	if err := s.CheckRiderAvailability(rider.Username, rider.Email, rider.Phone, rider.LicenseNumber); err != nil {
		return fmt.Errorf("%w: %v", ErrAvailabilityCheck, err)
	}

	// 加密密码
	if rider.PasswordHash != "" {
		hashedPassword, err := crypto.HashPassword(rider.PasswordHash)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPasswordHashing, err)
		}
		rider.PasswordHash = hashedPassword
	}

	// 创建配送员
	if err := s.riderRepo.Create(rider); err != nil {
		return fmt.Errorf("%w: %v", ErrDataSaveFailed, err)
	}

	s.logRiderRegistered(rider)
	return nil
}

// LoginRider 配送员登录
func (s *RiderService) LoginRider(loginInfo, password, loginType string) (string, error) {
	if loginInfo == "" || password == "" {
		return "", ErrLoginInfoEmpty
	}

	// 根据登录类型获取配送员信息
	rider, err := s.getRiderByLoginInfo(loginInfo, loginType)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrRiderNotFound, err)
	}

	// 验证密码
	if err := crypto.VerifyPassword(rider.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	// 检查配送员状态
	if !rider.IsActive {
		return "", ErrAccountDeactivated
	}

	// 生成JWT令牌
	token, err := s.jwtService.GenerateToken(rider.ID, "rider")
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	s.logRiderLogin(rider, loginType)
	return token, nil
}

// #endregion

// #region 配送员信息管理

// GetRiderByID 根据ID获取配送员信息
func (s *RiderService) GetRiderByID(id int64) (*model.Rider, error) {
	if id <= 0 {
		return nil, ErrInvalidRiderID
	}

	rider, err := s.riderRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRiderNotFound, err)
	}

	return rider, nil
}

// UpdateRiderProfile 更新配送员档案
func (s *RiderService) UpdateRiderProfile(riderID int64, name, vehicleType, vehicleNumber, licenseNumber string) error {
	if riderID <= 0 {
		return ErrInvalidRiderID
	}

	// 获取配送员信息
	rider, err := s.riderRepo.GetByID(riderID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRiderNotFound, err)
	}

	// 验证新的配送员信息
	if err := s.validateRiderFields(name, vehicleType, vehicleNumber, licenseNumber); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 更新配送员信息
	if err := rider.UpdateProfile(name, vehicleType, vehicleNumber, licenseNumber); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	if err := s.riderRepo.Update(rider); err != nil {
		return fmt.Errorf("%w: %v", ErrDataSaveFailed, err)
	}

	s.logRiderProfileUpdated(riderID)
	return nil
}

// UpdateRiderPassword 更新配送员密码
func (s *RiderService) UpdateRiderPassword(riderID int64, oldPassword, newPassword string) error {
	if riderID <= 0 {
		return ErrInvalidRiderID
	}
	if oldPassword == "" || newPassword == "" {
		return ErrPasswordsEmpty
	}

	// 获取配送员信息
	rider, err := s.riderRepo.GetByID(riderID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRiderNotFound, err)
	}

	// 验证旧密码
	if err := crypto.VerifyPassword(rider.PasswordHash, oldPassword); err != nil {
		return ErrOldPasswordIncorrect
	}

	// 加密新密码
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPasswordHashing, err)
	}

	// 更新密码
	rider.PasswordHash = hashedPassword
	if err := s.riderRepo.Update(rider); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logRiderPasswordUpdated(riderID)
	return nil
}

// #endregion

// #region 位置管理

// UpdateLocation 更新配送员位置
func (s *RiderService) UpdateLocation(riderID int64, lat, lng float64) error {
	if riderID <= 0 {
		return ErrInvalidRiderID
	}

	// 获取配送员信息以验证存在性
	rider, err := s.riderRepo.GetByID(riderID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRiderNotFound, err)
	}

	// 验证位置并更新
	if err := rider.UpdateLocation(lat, lng); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLocation, err)
	}

	// 保存到数据库
	if err := s.riderRepo.UpdateLocation(riderID, lat, lng); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logLocationUpdated(riderID, lat, lng)
	return nil
}

// GetRidersNearLocation 获取指定位置附近的配送员
func (s *RiderService) GetRidersNearLocation(lat, lng, radiusKm float64) ([]*model.Rider, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, ErrInvalidLocation
	}
	if radiusKm <= 0 {
		return nil, ErrRadiusInvalid
	}

	riders, err := s.riderRepo.GetRidersNearLocation(lat, lng, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetRiderList, err)
	}

	return riders, nil
}

// GetRidersByRegion 根据地理边界获取配送员
func (s *RiderService) GetRidersByRegion(bounds map[string]float64) ([]*model.Rider, error) {
	if len(bounds) == 0 {
		return nil, ErrBoundsEmpty
	}

	riders, err := s.riderRepo.GetRidersByRegion(bounds)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetRiderList, err)
	}

	return riders, nil
}

// #endregion

// #region 状态管理

// SetOnlineStatus 设置在线状态
func (s *RiderService) SetOnlineStatus(riderID int64, isOnline bool) error {
	if riderID <= 0 {
		return ErrInvalidRiderID
	}

	// 获取配送员信息
	rider, err := s.riderRepo.GetByID(riderID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRiderNotFound, err)
	}

	// 检查配送员状态
	if isOnline && !rider.IsActive {
		return ErrCannotSetInactiveOnline
	}

	// 更新在线状态
	if err := s.riderRepo.UpdateOnlineStatus(riderID, isOnline); err != nil {
		return fmt.Errorf("%w: %v", ErrDataUpdateFailed, err)
	}

	s.logStatusChanged(riderID, isOnline)
	return nil
}

// GetOnlineRiders 获取在线配送员列表
func (s *RiderService) GetOnlineRiders(offset, limit int) ([]*model.Rider, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	riders, total, err := s.riderRepo.GetOnlineRiders(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// GetActiveRiders 获取活跃配送员列表
func (s *RiderService) GetActiveRiders(offset, limit int) ([]*model.Rider, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	riders, total, err := s.riderRepo.GetActiveRiders(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// GetAvailableRiders 获取可接单的配送员
func (s *RiderService) GetAvailableRiders(lat, lng, radiusKm float64) ([]*model.Rider, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, ErrInvalidLocation
	}
	if radiusKm <= 0 {
		return nil, ErrRadiusInvalid
	}

	riders, err := s.riderRepo.GetAvailableRiders(lat, lng, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetRiderList, err)
	}

	return riders, nil
}

// #endregion

// #region 配送员验证

// ValidateRiderData 验证配送员数据
func (s *RiderService) ValidateRiderData(rider *model.Rider) error {
	if rider == nil {
		return ErrRiderNil
	}

	return rider.ValidateAll()
}

// CheckRiderAvailability 检查配送员可用性
func (s *RiderService) CheckRiderAvailability(username, email, phone, licenseNumber string) error {
	exists, err := s.riderRepo.CheckRiderExists(username, email, phone, licenseNumber)
	if err != nil {
		return ErrAvailabilityCheck
	}

	if exists {
		return ErrRiderAlreadyExists
	}

	return nil
}

// #endregion

// #region 配送员列表和搜索

// GetRiderList 获取配送员列表
func (s *RiderService) GetRiderList(offset, limit int) ([]*model.Rider, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	riders, total, err := s.riderRepo.GetRiderList(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// SearchRiders 搜索配送员
func (s *RiderService) SearchRiders(keyword string, offset, limit int) ([]*model.Rider, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	riders, total, err := s.riderRepo.SearchRiders(keyword, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// GetRidersByVehicleType 根据交通工具类型获取配送员
func (s *RiderService) GetRidersByVehicleType(vehicleType string, offset, limit int) ([]*model.Rider, int64, error) {
	if vehicleType == "" {
		return nil, 0, ErrVehicleTypeEmpty
	}
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	riders, total, err := s.riderRepo.GetRidersByVehicleType(vehicleType, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return riders, total, nil
}

// #endregion

// #region 配送员统计

// GetRiderStats 获取配送员统计信息
func (s *RiderService) GetRiderStats() (map[string]interface{}, error) {
	stats, err := s.riderRepo.GetRiderStats()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTopRidersByRating 获取评分最高的配送员
func (s *RiderService) GetTopRidersByRating(limit int) ([]*model.Rider, error) {
	if limit <= 0 {
		return nil, ErrLimitInvalid
	}

	riders, err := s.riderRepo.GetTopRidersByRating(limit)
	if err != nil {
		return nil, err
	}

	return riders, nil
}

// GetRidersByOrderCount 根据订单数量范围获取配送员
func (s *RiderService) GetRidersByOrderCount(minOrders, maxOrders int64) ([]*model.Rider, error) {
	if minOrders < 0 || maxOrders < minOrders {
		return nil, ErrOrderCountRangeInvalid
	}

	riders, err := s.riderRepo.GetRidersByOrderCount(minOrders, maxOrders)
	if err != nil {
		return nil, err
	}

	return riders, nil
}

// #endregion

// #region 私有辅助方法

// getRiderByLoginInfo 根据登录信息获取配送员
func (s *RiderService) getRiderByLoginInfo(loginInfo, loginType string) (*model.Rider, error) {
	switch loginType {
	case "email":
		return s.riderRepo.GetByEmail(loginInfo)
	case "phone":
		return s.riderRepo.GetByPhone(loginInfo)
	default:
		// 智能检测登录类型
		if validator.IsEmail(loginInfo) {
			return s.riderRepo.GetByEmail(loginInfo)
		} else if validator.IsPhone(loginInfo) {
			return s.riderRepo.GetByPhone(loginInfo)
		} else {
			return s.riderRepo.GetByUsername(loginInfo)
		}
	}
}

// validateRiderFields 验证配送员字段
func (s *RiderService) validateRiderFields(name, vehicleType, vehicleNumber, licenseNumber string) error {
	// 创建临时配送员对象进行验证
	tempRider := &model.Rider{
		Name:          name,
		VehicleType:   vehicleType,
		VehicleNumber: vehicleNumber,
		LicenseNumber: licenseNumber,
	}

	return tempRider.ValidateAll()
}

// #endregion

// #region 日志记录方法

// logRiderRegistered 记录配送员注册日志
func (s *RiderService) logRiderRegistered(rider *model.Rider) {
	log.Printf("配送员注册成功 - 用户名: %s, 邮箱: %s, 交通工具: %s, 时间: %s",
		rider.Username, rider.Email, rider.VehicleType, time.Now().Format("2006-01-02 15:04:05"))
}

// logRiderLogin 记录配送员登录日志
func (s *RiderService) logRiderLogin(rider *model.Rider, loginType string) {
	log.Printf("配送员登录成功 - 配送员ID: %d, 用户名: %s, 登录方式: %s, 时间: %s",
		rider.ID, rider.Username, loginType, time.Now().Format("2006-01-02 15:04:05"))
}

// logRiderProfileUpdated 记录配送员档案更新日志
func (s *RiderService) logRiderProfileUpdated(riderID int64) {
	log.Printf("配送员档案更新 - 配送员ID: %d, 时间: %s",
		riderID, time.Now().Format("2006-01-02 15:04:05"))
}

// logRiderPasswordUpdated 记录配送员密码更新日志
func (s *RiderService) logRiderPasswordUpdated(riderID int64) {
	log.Printf("配送员密码更新 - 配送员ID: %d, 时间: %s",
		riderID, time.Now().Format("2006-01-02 15:04:05"))
}

// logLocationUpdated 记录位置更新日志
func (s *RiderService) logLocationUpdated(riderID int64, lat, lng float64) {
	log.Printf("配送员位置更新 - 配送员ID: %d, 位置: (%.6f, %.6f), 时间: %s",
		riderID, lat, lng, time.Now().Format("2006-01-02 15:04:05"))
}

// logStatusChanged 记录状态变更日志
func (s *RiderService) logStatusChanged(riderID int64, isOnline bool) {
	status := "下线"
	if isOnline {
		status = "上线"
	}
	log.Printf("配送员状态变更 - 配送员ID: %d, 状态: %s, 时间: %s",
		riderID, status, time.Now().Format("2006-01-02 15:04:05"))
}

// #endregion
