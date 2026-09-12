package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/niaga-labs/niaga-labs-pet-lib-common/auth"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/middleware"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/response"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/application"
)

// AdminHandler handles admin HTTP requests for user management.
type AdminHandler struct {
	service *application.AuthService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(service *application.AuthService) *AdminHandler {
	return &AdminHandler{service: service}
}

// RegisterRoutes registers admin routes.
func (h *AdminHandler) RegisterRoutes(r *gin.RouterGroup, jwtManager *auth.JWTManager) {
	authMW := middleware.AuthMiddleware(jwtManager)
	adminRole := middleware.RequireRole(auth.RoleAdmin)

	admin := r.Group("/api/v1/admin")
	admin.Use(authMW, adminRole)
	{
		admin.GET("/users", h.ListUsers)
		admin.GET("/users/:id", h.GetUser)
		admin.POST("/users/:id/ban", h.BanUser)
		admin.GET("/stats/users", h.UserStats)
	}
}

// ListUsers handles GET /api/v1/admin/users.
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := h.service.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, users, total, page, limit)
}

// GetUser handles GET /api/v1/admin/users/:id.
func (h *AdminHandler) GetUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// BanUser handles POST /api/v1/admin/users/:id/ban.
func (h *AdminHandler) BanUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	if err := h.service.BanUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "user banned successfully"})
}

// UserStats handles GET /api/v1/admin/stats/users.
func (h *AdminHandler) UserStats(c *gin.Context) {
	stats, err := h.service.GetUserStats(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, stats)
}
