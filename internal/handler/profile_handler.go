package handler

import (
	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/middleware"
	"github.com/Kilat-Pet-Delivery/lib-common/response"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/application"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProfileHandler handles authenticated profile routes.
type ProfileHandler struct {
	service *application.ProfileService
	logger  *zap.Logger
}

// NewProfileHandler creates a profile handler.
func NewProfileHandler(service *application.ProfileService, logger *zap.Logger) *ProfileHandler {
	return &ProfileHandler{service: service, logger: logger}
}

// RegisterRoutes registers /me profile routes.
func (h *ProfileHandler) RegisterRoutes(r *gin.RouterGroup, jwtManager *auth.JWTManager) {
	protected := r.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		protected.GET("/me", h.GetProfile)
		protected.PUT("/me", h.UpdateProfile)
		protected.POST("/me/photo", h.UploadPhoto)
	}
}

// GetProfile handles GET /api/v1/me.
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	result, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("get profile failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateProfile handles PUT /api/v1/me.
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	var req application.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("update profile failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// UploadPhoto handles POST /api/v1/me/photo.
func (h *ProfileHandler) UploadPhoto(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	reader, err := file.Open()
	if err != nil {
		response.BadRequest(c, "file is invalid")
		return
	}
	defer reader.Close()

	result, err := h.service.UploadProfilePhoto(c.Request.Context(), userID, file.Filename, reader)
	if err != nil {
		h.logger.Error("upload profile photo failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
