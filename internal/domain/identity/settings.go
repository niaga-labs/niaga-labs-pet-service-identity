package identity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	LanguageEnglish = "en"
	LanguageMalay   = "ms"
	LanguageChinese = "zh"

	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"
)

// UserSettings stores user-level app preferences.
type UserSettings struct {
	userID               uuid.UUID
	notificationsEnabled map[string]bool
	language             string
	theme                string
	createdAt            time.Time
	updatedAt            time.Time
}

// DefaultSettings returns the implicit preferences for users without a row yet.
func DefaultSettings(userID uuid.UUID) *UserSettings {
	now := time.Now().UTC()
	return &UserSettings{
		userID: userID,
		notificationsEnabled: map[string]bool{
			"push":  true,
			"email": true,
			"sms":   true,
		},
		language:  LanguageEnglish,
		theme:     ThemeSystem,
		createdAt: now,
		updatedAt: now,
	}
}

// NewUserSettings validates and creates settings.
func NewUserSettings(userID uuid.UUID, notifications map[string]bool, language, theme string) (*UserSettings, error) {
	settings := DefaultSettings(userID)
	if err := settings.Update(notifications, language, theme); err != nil {
		return nil, err
	}
	return settings, nil
}

// ReconstructUserSettings rebuilds settings from persistence data.
func ReconstructUserSettings(userID uuid.UUID, notifications map[string]bool, language, theme string, createdAt, updatedAt time.Time) *UserSettings {
	return &UserSettings{
		userID:               userID,
		notificationsEnabled: cloneBoolMap(notifications),
		language:             language,
		theme:                theme,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
	}
}

func (s *UserSettings) UserID() uuid.UUID { return s.userID }
func (s *UserSettings) NotificationsEnabled() map[string]bool {
	return cloneBoolMap(s.notificationsEnabled)
}
func (s *UserSettings) Language() string     { return s.language }
func (s *UserSettings) Theme() string        { return s.theme }
func (s *UserSettings) CreatedAt() time.Time { return s.createdAt }
func (s *UserSettings) UpdatedAt() time.Time { return s.updatedAt }

// Update applies partial settings changes.
func (s *UserSettings) Update(notifications map[string]bool, language, theme string) error {
	if notifications != nil {
		s.notificationsEnabled = cloneBoolMap(notifications)
	}
	if language != "" {
		if !validLanguage(language) {
			return fmt.Errorf("language must be one of: en, ms, zh")
		}
		s.language = language
	}
	if theme != "" {
		if !validTheme(theme) {
			return fmt.Errorf("theme must be one of: system, light, dark")
		}
		s.theme = theme
	}
	s.updatedAt = time.Now().UTC()
	return nil
}

func validLanguage(language string) bool {
	return language == LanguageEnglish || language == LanguageMalay || language == LanguageChinese
}

func validTheme(theme string) bool {
	return theme == ThemeSystem || theme == ThemeLight || theme == ThemeDark
}

func cloneBoolMap(values map[string]bool) map[string]bool {
	if values == nil {
		return map[string]bool{}
	}
	clone := make(map[string]bool, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}
