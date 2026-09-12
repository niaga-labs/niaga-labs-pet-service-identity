package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/response"
	"github.com/niaga-labs/niaga-labs-pet-lib-proto/dto"
	"go.uber.org/zap"
)

// ForgotPasswordService defines the application-layer contract the handler depends on.
type ForgotPasswordService interface {
	ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error
}

// ForgotPasswordHandler handles POST /auth/forgot-password.
type ForgotPasswordHandler struct {
	service ForgotPasswordService
	logger  *zap.Logger
}

// NewForgotPasswordHandler creates a new ForgotPasswordHandler.
func NewForgotPasswordHandler(service ForgotPasswordService, logger *zap.Logger) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers forgot-password routes on the given router group.
func (h *ForgotPasswordHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.Group("/auth").POST("/forgot-password", h.ForgotPassword)
}

// ForgotPassword handles POST /auth/forgot-password.
// Always responds 202 Accepted regardless of whether the email matches a user.
func (h *ForgotPasswordHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.ForgotPassword(c.Request.Context(), req); err != nil {
		h.logger.Error("forgot password failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	c.Status(http.StatusAccepted)
}
