package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/niaga-labs/niaga-labs-pet-lib-common/auth"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/middleware"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/response"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/application"
)

// ReferralHandler handles HTTP requests for referral operations.
type ReferralHandler struct {
	service *application.ReferralService
}

// NewReferralHandler creates a new ReferralHandler.
func NewReferralHandler(service *application.ReferralService) *ReferralHandler {
	return &ReferralHandler{service: service}
}

// RegisterRoutes registers all referral routes.
func (h *ReferralHandler) RegisterRoutes(r *gin.RouterGroup, jwtManager *auth.JWTManager) {
	authMW := middleware.AuthMiddleware(jwtManager)

	referrals := r.Group("/api/v1/referrals")
	referrals.Use(authMW)
	{
		referrals.GET("/me", h.GetMyReferrals)
		referrals.GET("/code", h.GetMyReferralCode)
	}
}

// GetMyReferrals handles GET /api/v1/referrals/me.
func (h *ReferralHandler) GetMyReferrals(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.service.GetMyReferrals(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetMyReferralCode handles GET /api/v1/referrals/code.
func (h *ReferralHandler) GetMyReferralCode(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	code, err := h.service.GetOrCreateReferralCode(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"referral_code": code})
}
