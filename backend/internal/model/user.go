package model

import "time"

// 用户模型
type User struct {
	ID        		int64     	`json:"id" gorm:"primaryKey;autoIncrement;comment:用户ID"`
	Username  		string    	`json:"username" gorm:"unique;not null;size:50;comment:用户名"`
	PasswordHash	string    	`json:"password_hash" gorm:"not null;size:255;comment:用户密码"`
	Email     		string		`json:"email" gorm:"unique;not null;size:100;comment:用户邮箱"`
	Phone     		string		`json:"phone" gorm:"unique;not null;size:11;comment:用户手机号"`
	CreatedAt 		time.Time 	`json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt 		time.Time 	`json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	AvatarURL 		string    	`json:"avatar_url" gorm:"size:255;comment:用户头像URL"`
	// IsActive  	bool      	`json:"is_active" gorm:"default:true;comment:用户是否激活"`
	// IsAdmin   	bool      	`json:"is_admin" gorm:"default:false;comment:用户是否为管理员"`
}