package repository

import (
	"gorm.io/gorm"

	"github.com/Hermitf/the-pass/internal/model"
)

// 用户仓库
type UserRepository struct {
	db *gorm.DB // 数据库连接
}

// 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// 创建用户
func (r *UserRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

// 更新用户
func (r *UserRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

// 根据用户名获取用户
func (r *UserRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据邮箱获取用户
func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据手机号获取用户
func (r *UserRepository) GetByPhone(phone string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
