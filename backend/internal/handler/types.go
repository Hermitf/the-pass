package handler

// ================================================================
// 请求类型 - 用于API输入层
// ================================================================

// LoginRequest - 用户登录请求结构
type LoginRequest struct {
	LoginInfo string `json:"login_info" binding:"required" example:"user@example.com"`
	Password  string `json:"password" binding:"required" example:"password123"`
	LoginType string `json:"login_type" example:"password"` // "password" 或 "sms"
}

// RegisterRequest - 用户注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"new_user"`
	Password string `json:"password" binding:"required" example:"password123"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Phone    string `json:"phone" binding:"required" example:"1234567890"`
}

// ProfileRequest - 更新用户档案请求结构
type ProfileRequest struct {
	Username string `json:"username" binding:"required" example:"updated_user"`
	Email    string `json:"email" binding:"required,email" example:"updated_user@example.com"`
	Phone    string `json:"phone" binding:"required" example:"1234567890"`
}

// RiderLocationUpdateRequest - 更新配送员位置请求
type RiderLocationUpdateRequest struct {
	Latitude  float64 `json:"latitude" binding:"required" example:"39.9042"`
	Longitude float64 `json:"longitude" binding:"required" example:"116.4074"`
}

// RiderOnlineStatusRequest - 更新配送员在线状态请求
type RiderOnlineStatusRequest struct {
	IsOnline bool `json:"is_online" binding:"required" example:"true"`
}

// ================================================================
// 响应类型 - 用于API输出层
// ================================================================

// LoginResponse - 登录响应结构
type LoginResponse struct {
	Token   string `json:"token" example:"jwt_token_here"`
	Message string `json:"message" example:"登录成功"`
}

// RegisterResponse - 注册响应结构
type RegisterResponse struct {
	ID      int64  `json:"id,omitempty" example:"123"`
	Message string `json:"message" example:"注册成功"`
}

// ErrorResponse - 错误响应结构
type ErrorResponse struct {
	Error string `json:"error" example:"请求参数无效"`
}

// ================================================================
// 业务特定响应类型 - 用于领域场景
// ================================================================

// EmployeeListItemResponse - 员工列表项响应结构（商家查看员工列表时使用）
type EmployeeListItemResponse struct {
	ID       int64  `json:"id" example:"123"`
	Username string `json:"username" example:"employee123"`
	Name     string `json:"name" example:"员工张三"`
	Phone    string `json:"phone" example:"1234567890"`
	IsActive bool   `json:"is_active" example:"true"`
}
