package handler

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/service"
	"github.com/Hermitf/the-pass/pkg/crypto"
	"github.com/Hermitf/the-pass/pkg/sms"
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
func (h *AuthHandler) registerUserByType(ctx context.Context, userType string, registerReq *RegisterRequest, passwordHash string) error {
	switch userType {
	case "user":
		user := &model.User{
			Username:     registerReq.Username,
			PasswordHash: passwordHash,
			Email:        registerReq.Email,
			Phone:        registerReq.Phone,
		}
		return h.deps.UserService.RegisterUser(ctx, user, registerReq.SMSCode)

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
// TODO: 风控与审计日志待补充。
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

		err = h.registerUserByType(c.Request.Context(), userType, registerReq, passwordHash)
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
	if errors.Is(err, service.ErrInvalidCredentials) ||
		errors.Is(err, service.ErrInvalidPassword) ||
		errors.Is(err, service.ErrSMSCodeInvalid) ||
		errors.Is(err, service.ErrUnsupportedLoginType) {
		Unauthorized(c, "用户名或密码错误")
		return
	}
	InternalServerError(c, ErrMsgInternalServer, err.Error())
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
// TODO: 支持扫码登录并通过移动端进行二次确认。
// TODO: 引入登录失败次数限制、设备指纹识别等安全策略。
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

// #region SMS Code Endpoints

type sendSMSRequest struct {
	Phone string `json:"phone"`
}

type verifySMSRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// smsServiceContract 定义短信验证码服务需要满足的行为
type smsServiceContract interface {
	SendSMSCode(ctx context.Context, phone string) error
	VerifySMSCode(ctx context.Context, phone, code string) error
	CanSendSMSCode(ctx context.Context, phone string) (bool, time.Duration, error)
}

// handleSendSMS 通用发送验证码逻辑（支持请求上下文取消）
func handleSendSMS(c *gin.Context, svc smsServiceContract) {
	var req sendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Phone == "" {
		BadRequest(c, ErrMsgInvalidRequest, "手机号不能为空")
		return
	}
	if err := svc.SendSMSCode(c.Request.Context(), req.Phone); err != nil {
		BadRequest(c, "发送验证码失败", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

// handleVerifySMS 通用验证码校验逻辑
func handleVerifySMS(c *gin.Context, svc smsServiceContract) {
	var req verifySMSRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Phone == "" || req.Code == "" {
		BadRequest(c, ErrMsgInvalidRequest, "手机号或验证码不能为空")
		return
	}
	if err := svc.VerifySMSCode(c.Request.Context(), req.Phone, req.Code); err != nil {
		BadRequest(c, "验证码校验失败", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "验证成功"})
}

type canSendResponse struct {
	Allowed           bool   `json:"allowed"`
	RetryAfterSeconds int    `json:"retry_after_seconds"`
	Reason            string `json:"reason,omitempty"`
	Message           string `json:"message,omitempty"`
}

func handleCanSendSMS(c *gin.Context, svc smsServiceContract) {
	var req sendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Phone == "" {
		BadRequest(c, ErrMsgInvalidRequest, "手机号不能为空")
		return
	}
	allowed, retryAfter, err := svc.CanSendSMSCode(c.Request.Context(), req.Phone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPhoneInvalid):
			BadRequest(c, ErrMsgInvalidRequest, err.Error())
			return
		case errors.Is(err, sms.ErrSendTooFrequent):
			respondCanSend(c, allowed, retryAfter, "rate_limit", "发送过于频繁，请稍后再试")
			return
		case errors.Is(err, sms.ErrDailyLimitReached):
			respondCanSend(c, allowed, retryAfter, "daily_limit", "当天验证码发送次数已达上限")
			return
		case errors.Is(err, sms.ErrProviderDisabled):
			respondCanSend(c, allowed, retryAfter, "provider_disabled", "短信服务暂未启用")
			return
		default:
			InternalServerError(c, ErrMsgInternalServer, err.Error())
			return
		}
	}
	respondCanSend(c, true, 0, "", "可发送验证码")
}

func respondCanSend(c *gin.Context, allowed bool, retryAfter time.Duration, reason, message string) {
	retrySeconds := 0
	if retryAfter > 0 {
		retrySeconds = int(math.Ceil(retryAfter.Seconds()))
	}
	c.JSON(http.StatusOK, canSendResponse{
		Allowed:           allowed,
		RetryAfterSeconds: retrySeconds,
		Reason:            reason,
		Message:           message,
	})
}

// SendSMSCodeHandler 发送短信验证码
func (h *AuthHandler) SendSMSCodeHandler(c *gin.Context) {
	handleSendSMS(c, h.deps.UserService)
}

// VerifySMSCodeHandler 校验短信验证码
func (h *AuthHandler) VerifySMSCodeHandler(c *gin.Context) {
	handleVerifySMS(c, h.deps.UserService)
}

// CanSendSMSCodeHandler 用户验证码发送可用性检测
func (h *AuthHandler) CanSendSMSCodeHandler(c *gin.Context) {
	handleCanSendSMS(c, h.deps.UserService)
}

// SendMerchantSMSCodeHandler 商家发送短信验证码
func (h *AuthHandler) SendMerchantSMSCodeHandler(c *gin.Context) {
	handleSendSMS(c, h.deps.MerchantService)
}

// VerifyMerchantSMSCodeHandler 商家校验短信验证码
func (h *AuthHandler) VerifyMerchantSMSCodeHandler(c *gin.Context) {
	handleVerifySMS(c, h.deps.MerchantService)
}

// CanSendMerchantSMSCodeHandler 商家验证码发送可用性检测
func (h *AuthHandler) CanSendMerchantSMSCodeHandler(c *gin.Context) {
	handleCanSendSMS(c, h.deps.MerchantService)
}

// SendRiderSMSCodeHandler 配送员发送短信验证码
func (h *AuthHandler) SendRiderSMSCodeHandler(c *gin.Context) {
	handleSendSMS(c, h.deps.RiderService)
}

// VerifyRiderSMSCodeHandler 配送员校验短信验证码
func (h *AuthHandler) VerifyRiderSMSCodeHandler(c *gin.Context) {
	handleVerifySMS(c, h.deps.RiderService)
}

// CanSendRiderSMSCodeHandler 配送员验证码发送可用性检测
func (h *AuthHandler) CanSendRiderSMSCodeHandler(c *gin.Context) {
	handleCanSendSMS(c, h.deps.RiderService)
}

// #endregion

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
