package identity

import (
	"context"

	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/google/uuid"
)

// UserRepository defines persistence operations for User aggregates.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	ListAll(ctx context.Context, page, limit int) ([]*User, int64, error)
	ListByRole(ctx context.Context, role auth.UserRole, limit int) ([]*User, error)
	CountByRole(ctx context.Context) (map[string]int64, error)
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

// TokenRepository defines persistence operations for RefreshToken entities.
type TokenRepository interface {
	Save(ctx context.Context, token *RefreshToken) error
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

// PasswordResetRepository defines persistence operations for PasswordReset entities.
type PasswordResetRepository interface {
	Create(ctx context.Context, reset *PasswordReset) error
	FindByToken(ctx context.Context, token string) (*PasswordReset, error)
	FindAnyByToken(ctx context.Context, token string) (*PasswordReset, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	MarkUsedAndUpdatePassword(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID, newHash string) error
}

// RunnerApplicationRepository defines persistence operations for RunnerApplication entities.
type RunnerApplicationRepository interface {
	// Insert persists a new runner application and returns a formatted display ID
	// of the form KR-YYYY-NNNNN (e.g. KR-2026-00001).
	// The display ID is computed from the annual row count inside the same transaction
	// as the insert; a small race condition is possible under very high concurrency but
	// is acceptable for this use case.
	// Returns domain.NewAlreadyExistsError("RunnerApplication", "ic_number", icNumber)
	// if a row with the same ic_number already exists.
	Insert(ctx context.Context, app *RunnerApplication) (string, error)
}
