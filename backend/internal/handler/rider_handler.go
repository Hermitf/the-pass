package handler

import (
	"net/http"

	"github.com/Hermitf/the-pass/internal/service"
	"github.com/gin-gonic/gin"
)

// RiderHandlerDependencies contains all dependencies for RiderHandler
type RiderHandlerDependencies struct {
	RiderService service.RiderServiceInterface
}

// RiderHandler handles rider-specific business operations
// Note: Authentication operations (login/register/profile) are handled by AuthHandler
type RiderHandler struct {
	deps *RiderHandlerDependencies
}

// NewRiderHandler creates a new RiderHandler instance with dependency injection
func NewRiderHandler(riderService service.RiderServiceInterface) *RiderHandler {
	return &RiderHandler{
		deps: &RiderHandlerDependencies{
			RiderService: riderService,
		},
	}
}

// validateUserID validates the user ID from JWT context
func (h *RiderHandler) validateUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, ErrMsgUnauthorized)
		return 0, false
	}
	return userID.(int64), true
}

// getRiderAndRespond retrieves rider information and responds with it
func (h *RiderHandler) getRiderAndRespond(c *gin.Context, userID int64) {
	rider, err := h.deps.RiderService.GetRiderByID(userID)
	if err != nil {
		InternalServerError(c, ErrMsgInternalServer, err.Error())
		return
	}
	c.JSON(http.StatusOK, rider.ToResponse())
}

// validateOnlineStatusRequest validates online status update request
func (h *RiderHandler) validateOnlineStatusRequest(c *gin.Context) (*RiderOnlineStatusRequest, bool) {
	var statusReq RiderOnlineStatusRequest
	if err := c.ShouldBindJSON(&statusReq); err != nil {
		BadRequest(c, ErrMsgInvalidRequest, err.Error())
		return nil, false
	}
	return &statusReq, true
}

// UpdateOnlineStatusHandler handles updating rider online status
// @Summary Update rider online status
// @Description Update the online status of a rider
// @Tags riders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status body RiderOnlineStatusRequest true "Online status"
// @Success 200 {object} model.RiderResponse "Online status updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /riders/online-status [put]
func (h *RiderHandler) UpdateOnlineStatusHandler(c *gin.Context) {
	userID, valid := h.validateUserID(c)
	if !valid {
		return
	}

	statusReq, valid := h.validateOnlineStatusRequest(c)
	if !valid {
		return
	}

	err := h.deps.RiderService.SetOnlineStatus(userID, statusReq.IsOnline)
	if err != nil {
		InternalServerError(c, ErrMsgInternalServer, err.Error())
		return
	}

	h.getRiderAndRespond(c, userID)
}

// validateLocationRequest validates location update request
func (h *RiderHandler) validateLocationRequest(c *gin.Context) (*RiderLocationUpdateRequest, bool) {
	var locationReq RiderLocationUpdateRequest
	if err := c.ShouldBindJSON(&locationReq); err != nil {
		BadRequest(c, ErrMsgInvalidRequest, err.Error())
		return nil, false
	}
	return &locationReq, true
}

// UpdateLocationHandler handles updating rider location
// @Summary update rider location
// @Description Update the location of the logged-in rider
// @Tags riders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param location body RiderLocationUpdateRequest true "Location coordinates"
// @Success 200 {object} model.RiderResponse "Location updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /riders/location [put]
func (h *RiderHandler) UpdateLocationHandler(c *gin.Context) {
	userID, valid := h.validateUserID(c)
	if !valid {
		return
	}

	locationReq, valid := h.validateLocationRequest(c)
	if !valid {
		return
	}

	err := h.deps.RiderService.UpdateLocation(userID, locationReq.Latitude, locationReq.Longitude)
	if err != nil {
		InternalServerError(c, ErrMsgInternalServer, err.Error())
		return
	}

	h.getRiderAndRespond(c, userID)
}
