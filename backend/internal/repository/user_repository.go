package repository

import (
	"gorm.io/gorm"

	"github.com/Hermitf/the-pass/internal/model"
)

// #region 仓库定义

// UserRepositoryInterface 用户仓库接口 - 专注于数据访问层
type UserRepositoryInterface interface {
	// 基础CRUD操作
	CreateUser(user *model.User) error
	GetUserByID(id uint) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id uint) error

	// 查询方法
	GetUserByUsername(username string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByPhone(phone string) (*model.User, error)
	GetUserList(offset, limit int) ([]*model.User, int64, error)
	SearchUsers(keyword string, offset, limit int) ([]*model.User, int64, error)

	// 数据检查方法（纯数据层）
	ExistsWithUsername(username string) (bool, error)
	ExistsWithEmail(email string) (bool, error)
	ExistsWithPhone(phone string) (bool, error)
	CountUsers() (int64, error)
	CountUsersByStatus(status string) (int64, error)
}

// UserRepository 用户仓库实现
type UserRepository struct {
	db *gorm.DB
}

// #endregion

// #region 构造函数

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{
		db: db,
	}
}

// #endregion

// #region 基础CRUD操作

// CreateUser 创建用户
func (r *UserRepository) CreateUser(user *model.User) error {
	if user == nil {
		return ErrUserNil
	}

	return r.db.Create(user).Error
}

// GetUserByID 根据ID获取用户
func (r *UserRepository) GetUserByID(id uint) (*model.User, error) {
	if id == 0 {
		return nil, ErrUserIDZero
	}

	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户
func (r *UserRepository) UpdateUser(user *model.User) error {
	if user == nil {
		return ErrUserNil
	}

	return r.db.Save(user).Error
}

// DeleteUser 删除用户（软删除）
func (r *UserRepository) DeleteUser(id uint) error {
	if id == 0 {
		return ErrUserIDZero
	}

	return r.db.Delete(&model.User{}, id).Error
}

// #endregion

// #region 查询方法

// GetUserByUsername 根据用户名获取用户
func (r *UserRepository) GetUserByUsername(username string) (*model.User, error) {
	if username == "" {
		return nil, ErrUsernameEmpty
	}

	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, ErrEmailEmpty
	}

	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone 根据手机号获取用户
func (r *UserRepository) GetUserByPhone(phone string) (*model.User, error) {
	if phone == "" {
		return nil, ErrPhoneEmpty
	}

	var user model.User
	if err := r.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// #endregion

// #region 数据查询方法

// GetUserList 分页获取用户列表
func (r *UserRepository) GetUserList(offset, limit int) ([]*model.User, int64, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	var users []*model.User
	var total int64

	// 获取总数
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// SearchUsers 搜索用户
func (r *UserRepository) SearchUsers(keyword string, offset, limit int) ([]*model.User, int64, error) {
	if keyword == "" {
		return r.GetUserList(offset, limit)
	}

	if offset < 0 || limit <= 0 {
		return nil, 0, ErrPaginationInvalid
	}

	var users []*model.User
	var total int64

	searchPattern := "%" + keyword + "%"
	query := r.db.Model(&model.User{}).Where(
		"username LIKE ? OR email LIKE ? OR phone LIKE ? OR name LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)

	// 获取搜索结果总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询搜索结果
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// #endregion

// #region 数据检查方法

// ExistsWithUsername 检查用户名是否存在
func (r *UserRepository) ExistsWithUsername(username string) (bool, error) {
	if username == "" {
		return false, ErrUsernameEmpty
	}

	var count int64
	if err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsWithEmail 检查邮箱是否存在
func (r *UserRepository) ExistsWithEmail(email string) (bool, error) {
	if email == "" {
		return false, ErrEmailEmpty
	}

	var count int64
	if err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsWithPhone 检查手机号是否存在
func (r *UserRepository) ExistsWithPhone(phone string) (bool, error) {
	if phone == "" {
		return false, ErrPhoneEmpty
	}

	var count int64
	if err := r.db.Model(&model.User{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// CountUsers 获取用户总数
func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	if err := r.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountUsersByStatus 根据状态统计用户数
func (r *UserRepository) CountUsersByStatus(status string) (int64, error) {
	var count int64
	query := r.db.Model(&model.User{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// #endregion
