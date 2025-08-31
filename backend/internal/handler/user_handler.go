package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/service"
	"github.com/Hermitf/the-pass/internal/utils/userutils"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// LoginHandler handles user login
// @Summary user login
// @Description Login an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param login body LoginRequest true "User login information"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/login [post]
func (h *UserHandler) LoginHandler(c *gin.Context) {
	// logic for user login
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	// directly call the user service to handle login
	token, err := h.userService.LoginUser(loginReq.LoginInfo, loginReq.Password, loginReq.LoginType)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	c.JSON(http.StatusOK, LoginResponse{Token: token, Message: "Login successful"})
}

// RegisterHandler handles user registration
// @Summary user registration
// @Description Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration information"
// @Success 200 {object} RegisterResponse "Registration successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 409 {object} ErrorResponse "User already exists"
// @Router /users/register [post]
func (h *UserHandler) RegisterHandler(c *gin.Context) {
	// logic for user registration
	var registerReq RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}
	// create a new user model
	passwordHash, err := userutils.GeneratePasswordHash(registerReq.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Password hashing failed"})
		return
	}
	registerUser := &model.User{
		Username:     registerReq.Username,
		Email:        registerReq.Email,
		PasswordHash: passwordHash,
		Phone:        registerReq.Phone,
	}
	// 调用用户服务进行注册
	if err := h.userService.RegisterUser(registerUser); err != nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "User already exists"})
		return
	}
	c.JSON(http.StatusOK, RegisterResponse{Status: "Registration successful"})
}

// GetProfileHandler retrieves the user's profile
// @Summary Get user profile
// @Description Get the profile of the logged-in user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} UserProfileResponse "User profile retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/profile [get]
func (h *UserHandler) GetProfileHandler(c *gin.Context) {
	// get user ID from context
	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// retrieve user profile from the service
	user, err := h.userService.GetUserProfile(uint(userID.(int64)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
	})
}

// UpdateProfileHandler updates the user's profile
// @Summary Update user profile
// @Description Update the profile of the logged-in user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param profile body ProfileRequest true "User profile information"
// @Success 200 {object} UserProfileResponse "User profile updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfileHandler(c *gin.Context) {
	// get user ID from context
	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// parse the request body
	var profileReq ProfileRequest
	if err := c.ShouldBindJSON(&profileReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	// update user profile using the service
	if err := h.userService.UpdateUserProfile(uint(userID.(int64)), profileReq.Username, profileReq.Email, profileReq.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	// return the updated profile
	c.JSON(http.StatusOK, UserProfileResponse{
		ID:       userID.(int64),
		Username: profileReq.Username,
		Email:    profileReq.Email,
		Phone:    profileReq.Phone,
	})
}
