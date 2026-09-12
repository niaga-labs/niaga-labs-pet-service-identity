package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/auth"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/middleware"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/response"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/application"
	"go.uber.org/zap"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	service *application.AuthService
	logger  *zap.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(service *application.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all authentication routes on the given router group.
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup, jwtManager *auth.JWTManager) {
	authGroup := r.Group("/auth")
	{
		// Public routes (no authentication required)
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.RefreshToken)

		// Protected routes (authentication required)
		protected := authGroup.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			protected.POST("/logout", h.Logout)
			protected.GET("/profile", h.GetProfile)
			protected.PUT("/profile", h.UpdateProfile)
		}
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("registration failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Created(c, result)
}

// Login handles user authentication.
func (h *AuthHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// RefreshToken handles token refresh requests.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("token refresh failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// Logout handles user logout by revoking all refresh tokens.
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	if err := h.service.Logout(c.Request.Context(), userID); err != nil {
		h.logger.Error("logout failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "logged out successfully"})
}

// GetProfile retrieves the authenticated user's profile.
func (h *AuthHandler) GetProfile(c *gin.Context) {
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

// UpdateProfile updates the authenticated user's profile.
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
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
