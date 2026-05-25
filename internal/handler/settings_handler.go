package handler

import (
	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/middleware"
	"github.com/Kilat-Pet-Delivery/lib-common/response"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/application"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SettingsHandler handles authenticated settings routes.
type SettingsHandler struct {
	service *application.ProfileService
	logger  *zap.Logger
}

// NewSettingsHandler creates a settings handler.
func NewSettingsHandler(service *application.ProfileService, logger *zap.Logger) *SettingsHandler {
	return &SettingsHandler{service: service, logger: logger}
}

// RegisterRoutes registers /me/settings routes.
func (h *SettingsHandler) RegisterRoutes(r *gin.RouterGroup, jwtManager *auth.JWTManager) {
	protected := r.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		protected.GET("/me/settings", h.GetSettings)
		protected.PUT("/me/settings", h.UpdateSettings)
	}
}

// GetSettings handles GET /api/v1/me/settings.
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	result, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("get settings failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateSettings handles PUT /api/v1/me/settings.
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	var req application.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.UpdateSettings(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("update settings failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
