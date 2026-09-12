package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/response"
	"github.com/niaga-labs/niaga-labs-pet-lib-proto/dto"
	"go.uber.org/zap"
)

// ResetPasswordService defines the application-layer contract the reset-password handler depends on.
type ResetPasswordService interface {
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
}

// ResetPasswordHandler handles POST /auth/reset-password.
type ResetPasswordHandler struct {
	service ResetPasswordService
	logger  *zap.Logger
}

// NewResetPasswordHandler creates a new ResetPasswordHandler.
func NewResetPasswordHandler(service ResetPasswordService, logger *zap.Logger) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers reset-password routes on the given router group.
func (h *ResetPasswordHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.Group("/auth").POST("/reset-password", h.ResetPassword)
}

// ResetPassword handles POST /auth/reset-password.
func (h *ResetPasswordHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req); err != nil {
		h.logger.Warn("reset password failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "password reset successful"})
}
