package handler

import (
	"strconv"

	"github.com/Kilat-Pet-Delivery/lib-common/response"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/application"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AgentsHandler exposes support-agent lookup for incident assignment.
type AgentsHandler struct {
	service *application.ProfileService
	logger  *zap.Logger
}

// NewAgentsHandler creates an agents handler.
func NewAgentsHandler(service *application.ProfileService, logger *zap.Logger) *AgentsHandler {
	return &AgentsHandler{service: service, logger: logger}
}

// RegisterRoutes registers agent lookup routes.
func (h *AgentsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/agents/available", h.ListAvailable)
}

// ListAvailable handles GET /api/v1/agents/available.
func (h *AgentsHandler) ListAvailable(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	result, err := h.service.ListAvailableAgents(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("list available agents failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
