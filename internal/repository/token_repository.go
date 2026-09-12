package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/domain"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/identity"
	"gorm.io/gorm"
)

// RefreshTokenModel is the GORM model for the refresh_tokens table.
type RefreshTokenModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"type:text;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for GORM.
func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// toDomain converts a RefreshTokenModel to a domain RefreshToken.
func (m *RefreshTokenModel) toDomain() *identity.RefreshToken {
	return identity.ReconstructRefreshToken(
		m.ID,
		m.UserID,
		m.Token,
		m.ExpiresAt,
		m.Revoked,
		m.CreatedAt,
	)
}

// fromDomainRefreshToken converts a domain RefreshToken to a RefreshTokenModel.
func fromDomainRefreshToken(t *identity.RefreshToken) *RefreshTokenModel {
	return &RefreshTokenModel{
		ID:        t.ID(),
		UserID:    t.UserID(),
		Token:     t.Token(),
		ExpiresAt: t.ExpiresAt(),
		Revoked:   t.Revoked(),
		CreatedAt: t.CreatedAt(),
	}
}

// GormTokenRepository is a GORM-based implementation of TokenRepository.
type GormTokenRepository struct {
	db *gorm.DB
}

// NewGormTokenRepository creates a new GormTokenRepository.
func NewGormTokenRepository(db *gorm.DB) *GormTokenRepository {
	return &GormTokenRepository{db: db}
}

// Save persists a new refresh token to the database.
func (r *GormTokenRepository) Save(ctx context.Context, token *identity.RefreshToken) error {
	model := fromDomainRefreshToken(token)
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByToken retrieves a refresh token by its token string.
func (r *GormTokenRepository) FindByToken(ctx context.Context, token string) (*identity.RefreshToken, error) {
	var model RefreshTokenModel
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// RevokeAllForUser revokes all refresh tokens belonging to a specific user.
func (r *GormTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).
		Error
}
