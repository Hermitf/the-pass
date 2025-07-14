package handler


type LoginRequest struct {
	LoginInfo string `json:"login_info" binding:"required"`
    Password  string `json:"password" binding:"required"`
    LoginType string `json:"login_type"` // "password" 或 "sms"
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`	
}