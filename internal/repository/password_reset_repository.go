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

// PasswordResetModel is the GORM model for the password_resets table.
type PasswordResetModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Token     string     `gorm:"type:text;uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time `gorm:""`
	CreatedAt time.Time  `gorm:"not null;default:now()"`
}

// TableName specifies the table name for GORM.
func (PasswordResetModel) TableName() string {
	return "password_resets"
}

// toDomain converts a PasswordResetModel to a domain PasswordReset.
func (m *PasswordResetModel) toDomain() *identity.PasswordReset {
	return identity.ReconstructPasswordReset(
		m.ID,
		m.UserID,
		m.Token,
		m.ExpiresAt,
		m.UsedAt,
		m.CreatedAt,
	)
}

// fromDomainPasswordReset converts a domain PasswordReset to a PasswordResetModel.
func fromDomainPasswordReset(p *identity.PasswordReset) *PasswordResetModel {
	return &PasswordResetModel{
		ID:        p.ID(),
		UserID:    p.UserID(),
		Token:     p.Token(),
		ExpiresAt: p.ExpiresAt(),
		UsedAt:    p.UsedAt(),
		CreatedAt: p.CreatedAt(),
	}
}

// GormPasswordResetRepository is a GORM-based implementation of PasswordResetRepository.
type GormPasswordResetRepository struct {
	db *gorm.DB
}

// NewGormPasswordResetRepository creates a new GormPasswordResetRepository.
func NewGormPasswordResetRepository(db *gorm.DB) *GormPasswordResetRepository {
	return &GormPasswordResetRepository{db: db}
}

// Create persists a new password reset token to the database.
func (r *GormPasswordResetRepository) Create(ctx context.Context, reset *identity.PasswordReset) error {
	model := fromDomainPasswordReset(reset)
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByToken retrieves a non-expired password reset token by its token string.
// Returns domain.ErrNotFound if the token does not exist or has expired.
func (r *GormPasswordResetRepository) FindByToken(ctx context.Context, token string) (*identity.PasswordReset, error) {
	var model PasswordResetModel
	err := r.db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now().UTC()).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// FindAnyByToken retrieves a password reset token by its token string without filtering by expiry.
// Returns domain.ErrNotFound if the token does not exist at all.
func (r *GormPasswordResetRepository) FindAnyByToken(ctx context.Context, token string) (*identity.PasswordReset, error) {
	var model PasswordResetModel
	err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// MarkUsed sets the used_at timestamp on a password reset token.
// Returns domain.ErrNotFound if no rows were affected.
func (r *GormPasswordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&PasswordResetModel{}).
		Where("id = ?", id).
		Update("used_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// MarkUsedAndUpdatePassword atomically marks a reset token as used and updates the user's password hash.
// Both updates are wrapped in a single Postgres transaction; a failure in either rolls back both.
func (r *GormPasswordResetRepository) MarkUsedAndUpdatePassword(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID, newHash string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&PasswordResetModel{}).
			Where("id = ?", tokenID).
			Update("used_at", time.Now().UTC())
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}

		result = tx.Model(&UserModel{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"password_hash": newHash,
				"updated_at":    time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
}
