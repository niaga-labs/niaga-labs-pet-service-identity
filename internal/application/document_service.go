package application

import (
	"context"
	"fmt"
	"io"

	"github.com/Kilat-Pet-Delivery/lib-common/domain"
	"github.com/Kilat-Pet-Delivery/lib-common/storage"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/domain/identity"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DocumentDTO is the API shape for a verification document.
type DocumentDTO struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Kind             string     `json:"kind"`
	StorageURL       string     `json:"storage_url"`
	Status           string     `json:"status"`
	ReviewedByUserID *uuid.UUID `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt       *string    `json:"reviewed_at,omitempty"`
	CreatedAt        string     `json:"created_at"`
}

// ReviewDocumentRequest represents an admin document review decision.
type ReviewDocumentRequest struct {
	Status string `json:"status" binding:"required,oneof=verified rejected"`
}

// DocumentService handles verification document upload and review.
type DocumentService struct {
	documentRepo identity.DocumentRepository
	storage      storage.Storage
	logger       *zap.Logger
}

// NewDocumentService creates a document service.
func NewDocumentService(documentRepo identity.DocumentRepository, storage storage.Storage, logger *zap.Logger) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
		storage:      storage,
		logger:       logger,
	}
}

// UploadDocument stores a user document and starts it in pending status.
func (s *DocumentService) UploadDocument(ctx context.Context, userID uuid.UUID, kind, filename string, reader io.Reader) (*DocumentDTO, error) {
	if s.storage == nil {
		return nil, domain.NewValidationError("document storage is not configured")
	}
	if !identity.IsValidDocumentKind(kind) {
		return nil, domain.NewValidationError("document kind must be one of: ic, license, selfie, vehicle_reg")
	}

	key := fmt.Sprintf("identity/documents/%s/%s-%s%s", userID, uuid.NewString(), kind, safeExtension(filename))
	storageURL, err := s.storage.Upload(ctx, key, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	document, err := identity.NewUserDocument(userID, kind, storageURL)
	if err != nil {
		return nil, domain.NewValidationError(err.Error())
	}
	if err := s.documentRepo.Save(ctx, document); err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	result := toDocumentDTO(document)
	return &result, nil
}

// ListDocuments returns documents uploaded by a user.
func (s *DocumentService) ListDocuments(ctx context.Context, userID uuid.UUID) ([]DocumentDTO, error) {
	documents, err := s.documentRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	results := make([]DocumentDTO, len(documents))
	for i, document := range documents {
		results[i] = toDocumentDTO(document)
	}
	return results, nil
}

// ReviewDocument records an admin review decision.
func (s *DocumentService) ReviewDocument(ctx context.Context, documentID, reviewerID uuid.UUID, status string) (*DocumentDTO, error) {
	document, err := s.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		return nil, domain.NewNotFoundError("Document", documentID.String())
	}
	if err := document.Review(reviewerID, status); err != nil {
		return nil, domain.NewValidationError(err.Error())
	}
	if err := s.documentRepo.Update(ctx, document); err != nil {
		return nil, fmt.Errorf("failed to review document: %w", err)
	}

	result := toDocumentDTO(document)
	return &result, nil
}

func toDocumentDTO(document *identity.UserDocument) DocumentDTO {
	var reviewedAt *string
	if document.ReviewedAt() != nil {
		value := document.ReviewedAt().UTC().Format("2006-01-02T15:04:05Z07:00")
		reviewedAt = &value
	}
	return DocumentDTO{
		ID:               document.ID(),
		UserID:           document.UserID(),
		Kind:             document.Kind(),
		StorageURL:       document.StorageURL(),
		Status:           document.Status(),
		ReviewedByUserID: document.ReviewedByUserID(),
		ReviewedAt:       reviewedAt,
		CreatedAt:        document.CreatedAt().UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
