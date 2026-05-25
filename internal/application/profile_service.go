package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/domain"
	"github.com/Kilat-Pet-Delivery/lib-common/storage"
	"github.com/Kilat-Pet-Delivery/lib-proto/dto"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/domain/identity"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RoleSupportAgent auth.UserRole = "support_agent"

// SettingsDTO is the API shape for user app preferences.
type SettingsDTO struct {
	UserID               uuid.UUID       `json:"user_id"`
	NotificationsEnabled map[string]bool `json:"notifications_enabled"`
	Language             string          `json:"language"`
	Theme                string          `json:"theme"`
}

// UpdateSettingsRequest represents a partial settings update.
type UpdateSettingsRequest struct {
	NotificationsEnabled map[string]bool `json:"notifications_enabled"`
	Language             string          `json:"language"`
	Theme                string          `json:"theme"`
}

// ProfileService handles profile, settings, and agent-listing use cases.
type ProfileService struct {
	userRepo     identity.UserRepository
	settingsRepo identity.SettingsRepository
	storage      storage.Storage
	logger       *zap.Logger
}

// NewProfileService creates a profile service.
func NewProfileService(
	userRepo identity.UserRepository,
	settingsRepo identity.SettingsRepository,
	storage storage.Storage,
	logger *zap.Logger,
) *ProfileService {
	return &ProfileService{
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		storage:      storage,
		logger:       logger,
	}
}

// GetProfile retrieves a user profile.
func (s *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.NewNotFoundError("User", userID.String())
	}
	result := toUserDTO(user)
	return &result, nil
}

// UpdateProfile updates editable profile fields.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.NewNotFoundError("User", userID.String())
	}

	user.UpdateProfile(req.FullName, req.Phone, req.AvatarURL)
	user.IncrementVersion()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	result := toUserDTO(user)
	return &result, nil
}

// UploadProfilePhoto stores a new profile photo and records its URL on the user.
func (s *ProfileService) UploadProfilePhoto(ctx context.Context, userID uuid.UUID, filename string, reader io.Reader) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.NewNotFoundError("User", userID.String())
	}
	if s.storage == nil {
		return nil, domain.NewValidationError("profile photo storage is not configured")
	}

	key := fmt.Sprintf("identity/profile-photos/%s/%s%s", userID, uuid.NewString(), safeExtension(filename))
	photoURL, err := s.storage.Upload(ctx, key, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to upload profile photo: %w", err)
	}

	oldPhotoURL := user.AvatarURL()
	user.UpdateProfile("", "", photoURL)
	user.IncrementVersion()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile photo: %w", err)
	}

	if oldPhotoURL != "" {
		if err := s.storage.Delete(ctx, oldPhotoURL); err != nil && !errors.Is(err, storage.ErrObjectNotFound) {
			s.logger.Warn("failed to delete old profile photo", zap.Error(err), zap.String("user_id", userID.String()))
		}
	}

	result := toUserDTO(user)
	return &result, nil
}

// GetSettings returns persisted settings or the default settings.
func (s *ProfileService) GetSettings(ctx context.Context, userID uuid.UUID) (*SettingsDTO, error) {
	settings, err := s.settingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			settings = identity.DefaultSettings(userID)
		} else {
			return nil, fmt.Errorf("failed to get settings: %w", err)
		}
	}
	result := toSettingsDTO(settings)
	return &result, nil
}

// UpdateSettings persists user settings.
func (s *ProfileService) UpdateSettings(ctx context.Context, userID uuid.UUID, req UpdateSettingsRequest) (*SettingsDTO, error) {
	settings, err := s.settingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			settings = identity.DefaultSettings(userID)
		} else {
			return nil, fmt.Errorf("failed to get settings: %w", err)
		}
	}

	if err := settings.Update(req.NotificationsEnabled, req.Language, req.Theme); err != nil {
		return nil, domain.NewValidationError(err.Error())
	}
	if err := s.settingsRepo.Upsert(ctx, settings); err != nil {
		return nil, fmt.Errorf("failed to save settings: %w", err)
	}

	result := toSettingsDTO(settings)
	return &result, nil
}

// ListAvailableAgents returns verified support agents for incident assignment.
func (s *ProfileService) ListAvailableAgents(ctx context.Context, limit int) ([]dto.UserDTO, error) {
	users, err := s.userRepo.ListByRole(ctx, RoleSupportAgent, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list support agents: %w", err)
	}
	results := make([]dto.UserDTO, len(users))
	for i, user := range users {
		results[i] = toUserDTO(user)
	}
	return results, nil
}

func toSettingsDTO(settings *identity.UserSettings) SettingsDTO {
	return SettingsDTO{
		UserID:               settings.UserID(),
		NotificationsEnabled: settings.NotificationsEnabled(),
		Language:             settings.Language(),
		Theme:                settings.Theme(),
	}
}

func safeExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".heic":
		return ext
	default:
		return ""
	}
}
