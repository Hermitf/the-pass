package handler

import (
	"net/http"

	"github.com/Hermitf/the-pass/internal/model"
	"github.com/Hermitf/the-pass/internal/service"
	"github.com/gin-gonic/gin"
)

// MerchantHandlerDependencies contains all dependencies for MerchantHandler
type MerchantHandlerDependencies struct {
	MerchantService service.MerchantServiceInterface
	EmployeeService service.EmployeeServiceInterface
}

// MerchantHandler handles merchant-specific business operations
// Note: Authentication operations (login/register/profile) are handled by AuthHandler
type MerchantHandler struct {
	deps *MerchantHandlerDependencies
}

// NewMerchantHandler creates a new MerchantHandler instance with dependency injection
func NewMerchantHandler(merchantService service.MerchantServiceInterface, employeeService service.EmployeeServiceInterface) *MerchantHandler {
	return &MerchantHandler{
		deps: &MerchantHandlerDependencies{
			MerchantService: merchantService,
			EmployeeService: employeeService,
		},
	}
}

// validateMerchantID validates the merchant ID from JWT context
func (h *MerchantHandler) validateMerchantID(c *gin.Context) (int64, bool) {
	merchantID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, ErrMsgUnauthorized)
		return 0, false
	}
	return merchantID.(int64), true
}

// GetEmployeesHandler handles fetching merchant's employees
// @Summary get merchant employees
// @Description Get all employees of the logged-in merchant
// @Tags merchants
// @Produce json
// @Success 200 {array} model.EmployeeResponse "Employees list"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /merchants/employees [get]
func (h *MerchantHandler) GetEmployeesHandler(c *gin.Context) {
	merchantID, valid := h.validateMerchantID(c)
	if !valid {
		return
	}

	employees, err := h.deps.EmployeeService.GetEmployeesByMerchantID(merchantID)
	if err != nil {
		InternalServerError(c, ErrMsgInternalServer, err.Error())
		return
	}

	var employeeResponses []*model.EmployeeResponse
	for _, employee := range employees {
		employeeResponses = append(employeeResponses, employee.ToResponse())
	}

	c.JSON(http.StatusOK, employeeResponses)
}
