package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Kilat-Pet-Delivery/lib-common/domain"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/domain/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserDocumentModel is the GORM model for uploaded verification documents.
type UserDocumentModel struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	Kind             string     `gorm:"type:varchar(20);not null"`
	StorageURL       string     `gorm:"type:text;not null"`
	Status           string     `gorm:"type:varchar(20);not null"`
	ReviewedByUserID *uuid.UUID `gorm:"type:uuid;column:reviewed_by_user_id"`
	ReviewedAt       *time.Time `gorm:"column:reviewed_at"`
	CreatedAt        time.Time  `gorm:"not null;default:now()"`
}

func (UserDocumentModel) TableName() string {
	return "user_documents"
}

// GormDocumentRepository persists verification documents.
type GormDocumentRepository struct {
	db *gorm.DB
}

// NewGormDocumentRepository creates a document repository.
func NewGormDocumentRepository(db *gorm.DB) *GormDocumentRepository {
	return &GormDocumentRepository{db: db}
}

// Save persists a new document.
func (r *GormDocumentRepository) Save(ctx context.Context, document *identity.UserDocument) error {
	return r.db.WithContext(ctx).Create(documentFromDomain(document)).Error
}

// FindByID retrieves a document by ID.
func (r *GormDocumentRepository) FindByID(ctx context.Context, id uuid.UUID) (*identity.UserDocument, error) {
	var model UserDocumentModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// ListByUserID lists documents owned by a user.
func (r *GormDocumentRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*identity.UserDocument, error) {
	var models []UserDocumentModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	documents := make([]*identity.UserDocument, len(models))
	for i := range models {
		documents[i] = models[i].toDomain()
	}
	return documents, nil
}

// Update persists a document review decision.
func (r *GormDocumentRepository) Update(ctx context.Context, document *identity.UserDocument) error {
	result := r.db.WithContext(ctx).
		Model(&UserDocumentModel{}).
		Where("id = ?", document.ID()).
		Updates(documentFromDomain(document))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (m *UserDocumentModel) toDomain() *identity.UserDocument {
	return identity.ReconstructUserDocument(
		m.ID,
		m.UserID,
		m.Kind,
		m.StorageURL,
		m.Status,
		m.ReviewedByUserID,
		m.ReviewedAt,
		m.CreatedAt,
	)
}

func documentFromDomain(document *identity.UserDocument) *UserDocumentModel {
	return &UserDocumentModel{
		ID:               document.ID(),
		UserID:           document.UserID(),
		Kind:             document.Kind(),
		StorageURL:       document.StorageURL(),
		Status:           document.Status(),
		ReviewedByUserID: document.ReviewedByUserID(),
		ReviewedAt:       document.ReviewedAt(),
		CreatedAt:        document.CreatedAt(),
	}
}
