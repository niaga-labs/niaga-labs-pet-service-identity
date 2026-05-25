package handler

import (
	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/middleware"
	"github.com/Kilat-Pet-Delivery/lib-common/response"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/application"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DocumentsHandler handles verification document routes.
type DocumentsHandler struct {
	service *application.DocumentService
	logger  *zap.Logger
}

// NewDocumentsHandler creates a documents handler.
func NewDocumentsHandler(service *application.DocumentService, logger *zap.Logger) *DocumentsHandler {
	return &DocumentsHandler{service: service, logger: logger}
}

// RegisterRoutes registers user and admin document routes.
func (h *DocumentsHandler) RegisterRoutes(r *gin.RouterGroup, jwtManager *auth.JWTManager) {
	protected := r.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		protected.POST("/me/documents", h.UploadDocument)
		protected.GET("/me/documents", h.ListDocuments)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(jwtManager), middleware.RequireRole(auth.RoleAdmin))
	{
		admin.POST("/documents/:id/review", h.ReviewDocument)
	}
}

// UploadDocument handles POST /api/v1/me/documents.
func (h *DocumentsHandler) UploadDocument(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	kind := c.PostForm("kind")
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

	result, err := h.service.UploadDocument(c.Request.Context(), userID, kind, file.Filename, reader)
	if err != nil {
		h.logger.Error("upload document failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

// ListDocuments handles GET /api/v1/me/documents.
func (h *DocumentsHandler) ListDocuments(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	result, err := h.service.ListDocuments(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list documents failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// ReviewDocument handles POST /api/v1/admin/documents/:id/review.
func (h *DocumentsHandler) ReviewDocument(c *gin.Context) {
	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "user ID not found in context")
		return
	}

	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid document ID")
		return
	}

	var req application.ReviewDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.ReviewDocument(c.Request.Context(), documentID, reviewerID, req.Status)
	if err != nil {
		h.logger.Error("review document failed", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
