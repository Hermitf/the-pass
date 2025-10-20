package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// #region HTTP错误响应结构扩展

// DetailedErrorResponse 详细错误响应结构
type DetailedErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse HTTP成功响应结构
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Code    int         `json:"code"`
}

// #endregion

// #region 通用HTTP错误常量

const (
	ErrMsgInvalidRequest   = "请求参数无效"
	ErrMsgUnauthorized     = "未授权访问"
	ErrMsgForbidden        = "权限不足"
	ErrMsgNotFound         = "资源不存在"
	ErrMsgInternalServer   = "服务器内部错误"
	ErrMsgInvalidUserType  = "无效的用户类型"
	ErrMsgValidationFailed = "参数验证失败"
)

// #endregion

// #region 统一错误处理函数

// RespondWithError 统一错误响应
func RespondWithError(c *gin.Context, statusCode int, errorType, message string, details interface{}) {
	c.JSON(statusCode, DetailedErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    statusCode,
		Details: details,
	})
}

// RespondWithSuccess 统一成功响应
func RespondWithSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, SuccessResponse{
		Data:    data,
		Message: message,
		Code:    statusCode,
	})
}

// RespondWithSimpleError 简单错误响应（兼容现有ErrorResponse）
func RespondWithSimpleError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: message,
	})
}

// #endregion

// #region 常用错误响应函数

// BadRequest 400错误
func BadRequest(c *gin.Context, message string, details interface{}) {
	RespondWithError(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	RespondWithError(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	RespondWithError(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

// Conflict 409错误
func Conflict(c *gin.Context, message string, details interface{}) {
	RespondWithError(c, http.StatusConflict, "CONFLICT", message, details)
}

// UnprocessableEntity 422错误
func UnprocessableEntity(c *gin.Context, message string, details interface{}) {
	RespondWithError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message, details)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, message string, details interface{}) {
	RespondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, details)
}

// #endregion

// #region 业务错误映射

// ErrorMapping 错误映射配置
type ErrorMapping struct {
	StatusCode int
	ErrorType  string
	Message    string
}

// 错误映射表
var errorMappings = map[string]ErrorMapping{
	"用户已存在":  {http.StatusConflict, "CONFLICT", "用户已存在"},
	"用户不存在":  {http.StatusNotFound, "NOT_FOUND", "用户不存在"},
	"无效凭证":   {http.StatusUnauthorized, "UNAUTHORIZED", "用户名或密码错误"},
	"员工已存在":  {http.StatusConflict, "CONFLICT", "员工已存在"},
	"员工不存在":  {http.StatusNotFound, "NOT_FOUND", "员工不存在"},
	"商家已存在":  {http.StatusConflict, "CONFLICT", "商家已存在"},
	"商家不存在":  {http.StatusNotFound, "NOT_FOUND", "商家不存在"},
	"配送员已存在": {http.StatusConflict, "CONFLICT", "配送员已存在"},
	"配送员不存在": {http.StatusNotFound, "NOT_FOUND", "配送员不存在"},
}

// HandleServiceError 将service层错误映射为HTTP错误
func HandleServiceError(c *gin.Context, err error) {
	if mapping, exists := errorMappings[err.Error()]; exists {
		RespondWithError(c, mapping.StatusCode, mapping.ErrorType, mapping.Message, nil)
	} else {
		// 未知错误，返回500
		InternalServerError(c, "服务器内部错误", err.Error())
	}
}

// #endregion

// #region 参数验证错误处理

// ValidationError 参数验证错误详情
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// HandleValidationErrors 处理参数验证错误
func HandleValidationErrors(c *gin.Context, validationErrors []ValidationError) {
	UnprocessableEntity(c, "参数验证失败", validationErrors)
}

// #endregion
