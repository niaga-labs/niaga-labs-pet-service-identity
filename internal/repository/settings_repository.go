package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Kilat-Pet-Delivery/lib-common/domain"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/domain/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserSettingsModel is the GORM model for the user_settings table.
type UserSettingsModel struct {
	UserID               uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	NotificationsEnabled []byte    `gorm:"type:jsonb;not null;column:notifications_enabled"`
	Language             string    `gorm:"type:varchar(2);not null"`
	Theme                string    `gorm:"type:varchar(10);not null"`
	CreatedAt            time.Time `gorm:"not null;default:now()"`
	UpdatedAt            time.Time `gorm:"not null;default:now()"`
}

func (UserSettingsModel) TableName() string {
	return "user_settings"
}

// GormSettingsRepository persists user settings in Postgres.
type GormSettingsRepository struct {
	db *gorm.DB
}

// NewGormSettingsRepository creates a settings repository.
func NewGormSettingsRepository(db *gorm.DB) *GormSettingsRepository {
	return &GormSettingsRepository{db: db}
}

// FindByUserID returns persisted settings for a user.
func (r *GormSettingsRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*identity.UserSettings, error) {
	var model UserSettingsModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return settingsToDomain(&model)
}

// Upsert creates or replaces settings for a user.
func (r *GormSettingsRepository) Upsert(ctx context.Context, settings *identity.UserSettings) error {
	model, err := settingsFromDomain(settings)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func settingsToDomain(model *UserSettingsModel) (*identity.UserSettings, error) {
	notifications := map[string]bool{}
	if len(model.NotificationsEnabled) > 0 {
		if err := json.Unmarshal(model.NotificationsEnabled, &notifications); err != nil {
			return nil, err
		}
	}
	return identity.ReconstructUserSettings(
		model.UserID,
		notifications,
		model.Language,
		model.Theme,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

func settingsFromDomain(settings *identity.UserSettings) (*UserSettingsModel, error) {
	notifications, err := json.Marshal(settings.NotificationsEnabled())
	if err != nil {
		return nil, err
	}
	return &UserSettingsModel{
		UserID:               settings.UserID(),
		NotificationsEnabled: notifications,
		Language:             settings.Language(),
		Theme:                settings.Theme(),
		CreatedAt:            settings.CreatedAt(),
		UpdatedAt:            settings.UpdatedAt(),
	}, nil
}
