package identity

import (
	"context"

	"github.com/google/uuid"
)

// SettingsRepository persists user app preferences.
type SettingsRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*UserSettings, error)
	Upsert(ctx context.Context, settings *UserSettings) error
}

// DocumentRepository persists user verification documents.
type DocumentRepository interface {
	Save(ctx context.Context, document *UserDocument) error
	FindByID(ctx context.Context, id uuid.UUID) (*UserDocument, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*UserDocument, error)
	Update(ctx context.Context, document *UserDocument) error
}
