package handler

// ==== Request structures ====

// LoginRequest represents the request structure for user login
type LoginRequest struct {
	LoginInfo string `json:"login_info" binding:"required" example:"user@example.com"`
	Password  string `json:"password" binding:"required" example:"password123"`
	LoginType string `json:"login_type" example:"password"` // "password" 或 "sms"
}

// RegisterRequest represents the request structure for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"new_user"`
	Password string `json:"password" binding:"required" example:"password123"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Phone    string `json:"phone" binding:"required" example:"1234567890"`
}

// ProfileRequest represents the request structure for updating user profile
type ProfileRequest struct {
	Username string `json:"username" binding:"required" example:"updated_user"`
	Email    string `json:"email" binding:"required,email" example:"updated_user@example.com"`
	Phone    string `json:"phone" binding:"required" example:"1234567890"`
}

// ==== Response structures ====

// LoginResponse represents the response structure for user login
type LoginResponse struct {
	Token   string `json:"token" binding:"required" example:"jwt_token"`
	Message string `json:"message" example:"Login successful"`
}

// RegisterResponse represents the response structure for user registration
type RegisterResponse struct {
	Status string `json:"status" example:"Registration successful"`
}

// UserProfileResponse represents the response structure for user profile
type UserProfileResponse struct {
	ID       int64  `json:"id" example:"123"`
	Username string `json:"username" example:"user123"`
	Email    string `json:"email" example:"user@example.com"`
	Phone    string `json:"phone" example:"1234567890"`
}

// ==== Common Response structure ====

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request"`
}
