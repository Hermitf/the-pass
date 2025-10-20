package handler

import (
	"errors"
	"net/http"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/service"
	"github.com/Hermitf/the-pass/pkg/crypto"
	"github.com/gin-gonic/gin"
)

// @title The Pass API
// @version 1.0
// @description The Pass API documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:13544
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// #region Dependency Injection & Constructor

// AuthHandlerDependencies contains all dependencies for AuthHandler
type AuthHandlerDependencies struct {
	UserService     service.UserServiceInterface
	EmployeeService service.EmployeeServiceInterface
	MerchantService service.MerchantServiceInterface
	RiderService    service.RiderServiceInterface
}

// AuthHandler handles unified authentication for all user types
type AuthHandler struct {
	deps *AuthHandlerDependencies
}

func NewAuthHandler(userService service.UserServiceInterface, employeeService service.EmployeeServiceInterface, merchantService service.MerchantServiceInterface, riderService service.RiderServiceInterface) *AuthHandler {
	return &AuthHandler{
		deps: &AuthHandlerDependencies{
			UserService:     userService,
			EmployeeService: employeeService,
			MerchantService: merchantService,
			RiderService:    riderService,
		},
	}
}

// #endregion

// #region User Registration Module

// validateRegistrationRequest validates and parses registration request
func (h *AuthHandler) validateRegistrationRequest(c *gin.Context) (*RegisterRequest, string, error) {
	var registerReq RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		return nil, "", err
	}

	passwordHash, err := crypto.HashPassword(registerReq.Password)
	if err != nil {
		return nil, "", err
	}

	return &registerReq, passwordHash, nil
}

// registerUserByType handles registration for different user types
func (h *AuthHandler) registerUserByType(userType string, registerReq *RegisterRequest, passwordHash string) error {
	switch userType {
	case "user":
		user := &model.User{
			Username:     registerReq.Username,
			PasswordHash: passwordHash,
			Email:        registerReq.Email,
			Phone:        registerReq.Phone,
		}
		return h.deps.UserService.RegisterUser(user)

	case "employee":
		employee := &model.Employee{
			Username:     registerReq.Username,
			PasswordHash: passwordHash,
			Email:        registerReq.Email,
			Phone:        registerReq.Phone,
		}
		return h.deps.EmployeeService.RegisterEmployee(employee)

	case "merchant":
		merchant := &model.Merchant{
			Username:     registerReq.Username,
			PasswordHash: passwordHash,
			Email:        registerReq.Email,
			Phone:        registerReq.Phone,
		}
		return h.deps.MerchantService.RegisterMerchant(merchant)

	case "rider":
		rider := &model.Rider{
			Username:     registerReq.Username,
			PasswordHash: passwordHash,
			Email:        registerReq.Email,
			Phone:        registerReq.Phone,
		}
		return h.deps.RiderService.RegisterRider(rider)

	default:
		return errors.New(ErrMsgInvalidUserType)
	}
}

// handleRegistrationError handles registration errors and returns appropriate responses
func (h *AuthHandler) handleRegistrationError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrUserAlreadyExists) ||
		errors.Is(err, service.ErrEmployeeAlreadyExists) ||
		errors.Is(err, service.ErrMerchantAlreadyExists) ||
		errors.Is(err, service.ErrRiderAlreadyExists) {
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	} else {
		InternalServerError(c, ErrMsgInternalServer, err.Error())
	}
}

// RegisterHandler - common registration handler
// @Summary common registration interface
// @Description supports unified registration for users, employees, and merchants, distinguished by userType parameter
// @Tags Authentication
// @Accept json
// @Produce json
// @Param userType path string true "user type" Enums(user, employee, merchant)
// @Param registerRequest body RegisterRequest true "registration information"
// @Success 200 {object} RegisterResponse "registration successful"
// @Failure 400 {object} ErrorResponse "invalid request"
// @Failure 409 {object} ErrorResponse "user already exists"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /{userType}/register [post]
func (h *AuthHandler) RegisterHandler(userType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		registerReq, passwordHash, err := h.validateRegistrationRequest(c)
		if err != nil {
			BadRequest(c, ErrMsgInvalidRequest, err.Error())
			return
		}

		if userType != "user" && userType != "employee" && userType != "merchant" && userType != "rider" {
			BadRequest(c, ErrMsgInvalidUserType, nil)
			return
		}

		err = h.registerUserByType(userType, registerReq, passwordHash)
		if err != nil {
			h.handleRegistrationError(c, err)
			return
		}

		c.JSON(http.StatusOK, RegisterResponse{Message: "注册成功"})
	}
}

// #endregion

// #region User Login

// validateLoginRequest validates and parses login request
func (h *AuthHandler) validateLoginRequest(c *gin.Context) (*LoginRequest, error) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		return nil, err
	}
	return &loginReq, nil
}

// authenticateUserByType handles authentication for different user types
func (h *AuthHandler) authenticateUserByType(userType string, loginReq *LoginRequest) (string, error) {
	switch userType {
	case "user":
		return h.deps.UserService.LoginUser(loginReq.LoginInfo, loginReq.Password, loginReq.LoginType)
	case "employee":
		return h.deps.EmployeeService.LoginEmployee(loginReq.LoginInfo, loginReq.Password, loginReq.LoginType)
	case "merchant":
		return h.deps.MerchantService.LoginMerchant(loginReq.LoginInfo, loginReq.Password, loginReq.LoginType)
	case "rider":
		return h.deps.RiderService.LoginRider(loginReq.LoginInfo, loginReq.Password, loginReq.LoginType)
	default:
		return "", errors.New(ErrMsgInvalidUserType)
	}
}

// handleLoginError handles login errors and returns appropriate responses
func (h *AuthHandler) handleLoginError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidCredentials) {
		Unauthorized(c, "用户名或密码错误")
	} else {
		InternalServerError(c, ErrMsgInternalServer, err.Error())
	}
}

// LoginHandler - common login handler
// @Summary common login interface
// @Description supports unified login for users, employees, and merchants, distinguished by userType parameter
// @Tags Authentication
// @Accept json
// @Produce json
// @Param userType path string true "user type" Enums(user, employee, merchant)
// @Param loginRequest body LoginRequest true "login information"
// @Success 200 {object} LoginResponse "login successful"
// @Failure 400 {object} ErrorResponse "invalid request"
// @Failure 401 {object} ErrorResponse "unauthorized"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /{userType}/login [post]
func (h *AuthHandler) LoginHandler(userType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		loginReq, err := h.validateLoginRequest(c)
		if err != nil {
			BadRequest(c, ErrMsgInvalidRequest, err.Error())
			return
		}

		if userType != "user" && userType != "employee" && userType != "merchant" && userType != "rider" {
			BadRequest(c, ErrMsgInvalidUserType, nil)
			return
		}

		token, err := h.authenticateUserByType(userType, loginReq)
		if err != nil {
			h.handleLoginError(c, err)
			return
		}

		c.JSON(http.StatusOK, LoginResponse{Token: token, Message: "登录成功"})
	}
}

// #endregion

// #region User Profile Module

// getUserProfileByType retrieves user profile by type
func (h *AuthHandler) getUserProfileByType(userType string, userID int64) (interface{}, error) {
	switch userType {
	case "user":
		user, err := h.deps.UserService.GetUserByID(userID)
		if err != nil {
			return nil, err
		}
		return user.ToResponse(), nil

	case "employee":
		employee, err := h.deps.EmployeeService.GetEmployeeByID(userID)
		if err != nil {
			return nil, err
		}
		return employee.ToResponse(), nil

	case "merchant":
		merchant, err := h.deps.MerchantService.GetMerchantByID(userID)
		if err != nil {
			return nil, err
		}
		return merchant.ToResponse(), nil

	case "rider":
		rider, err := h.deps.RiderService.GetRiderByID(userID)
		if err != nil {
			return nil, err
		}
		return rider.ToResponse(), nil

	default:
		return nil, errors.New(ErrMsgInvalidUserType)
	}
}

// GetProfileHandler - get profile handler
// @Summary get user profile
// @Description retrieve the profile information of the currently logged-in user, supporting users, employees, and merchants
// @Tags Authentication
// @Accept json
// @Produce json
// @Param userType path string true "user type" Enums(user, employee, merchant, rider)
// @Security BearerAuth
// @Success 200 {object} model.UserResponse "user profile (when userType=user)"
// @Success 200 {object} model.EmployeeResponse "employee profile (when userType=employee)"
// @Success 200 {object} model.MerchantResponse "merchant profile (when userType=merchant)"
// @Success 200 {object} model.RiderResponse "rider profile (when userType=rider)"
// @Failure 401 {object} ErrorResponse "unauthorized"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /{userType}/profile [get]
func (h *AuthHandler) GetProfileHandler(userType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			Unauthorized(c, ErrMsgUnauthorized)
			return
		}

		if userType != "user" && userType != "employee" && userType != "merchant" && userType != "rider" {
			BadRequest(c, ErrMsgInvalidUserType, nil)
			return
		}

		profile, err := h.getUserProfileByType(userType, userID.(int64))
		if err != nil {
			InternalServerError(c, ErrMsgInternalServer, err.Error())
			return
		}

		c.JSON(http.StatusOK, profile)
	}
}

// #endregion

// #region Employee Management Module

// createEmployeeForMerchant creates an employee associated with the merchant
func (h *AuthHandler) createEmployeeForMerchant(addEmployeeReq *RegisterRequest, merchantID int64) (*model.Employee, error) {
	passwordHash, err := crypto.HashPassword(addEmployeeReq.Password)
	if err != nil {
		return nil, err
	}

	employee := &model.Employee{
		Username:     addEmployeeReq.Username,
		PasswordHash: passwordHash,
		Email:        addEmployeeReq.Email,
		Phone:        addEmployeeReq.Phone,
		MerchantID:   merchantID,
	}

	err = h.deps.EmployeeService.RegisterEmployee(employee)
	if err != nil {
		return nil, err
	}

	return employee, nil
}

// AddEmployeeHandler 商家端新增员工处理器
// @Summary 商家添加员工
// @Description 商家为自己的店铺添加员工账号
// @Tags Merchant Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param addEmployeeRequest body RegisterRequest true "员工信息"
// @Success 200 {object} RegisterResponse "员工添加成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 409 {object} ErrorResponse "员工已存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /merchants/employees [post]
func (h *AuthHandler) AddEmployeeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID, exists := c.Get("userID")
		if !exists {
			Unauthorized(c, ErrMsgUnauthorized)
			return
		}

		var addEmployeeReq RegisterRequest
		if err := c.ShouldBindJSON(&addEmployeeReq); err != nil {
			BadRequest(c, ErrMsgInvalidRequest, err.Error())
			return
		}

		employee, err := h.createEmployeeForMerchant(&addEmployeeReq, merchantID.(int64))
		if err != nil {
			if errors.Is(err, service.ErrEmployeeAlreadyExists) {
				Conflict(c, err.Error(), nil)
			} else {
				InternalServerError(c, ErrMsgInternalServer, err.Error())
			}
			return
		}

		c.JSON(http.StatusOK, RegisterResponse{ID: employee.ID, Message: "员工添加成功"})
	}
}

// #endregion
