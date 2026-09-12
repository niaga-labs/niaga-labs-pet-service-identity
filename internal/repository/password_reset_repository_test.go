//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/domain"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/identity"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5435 user=kilat password=kilat_secret dbname=kilat_identity sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	if err := db.Exec("TRUNCATE TABLE password_resets").Error; err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
	return db
}

func seedTestUser(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	model := repository.UserModel{
		ID:           userID,
		Email:        "resettest-" + userID.String() + "@kilat.my",
		Phone:        "",
		PasswordHash: "$2a$10$placeholder",
		FullName:     "Reset Test User",
		Role:         "runner",
	}
	if err := db.Create(&model).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	return userID
}

func TestPasswordResetRepo_CreateAndFindByToken(t *testing.T) {
	db := setupTestDB(t)
	userID := seedTestUser(t, db)

	repo := repository.NewGormPasswordResetRepository(db)
	ctx := context.Background()

	reset := identity.NewPasswordReset(userID, "valid-token-abc123", time.Now().UTC().Add(time.Hour))
	if err := repo.Create(ctx, reset); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByToken(ctx, "valid-token-abc123")
	if err != nil {
		t.Fatalf("FindByToken returned error: %v", err)
	}
	if found == nil {
		t.Fatal("FindByToken returned nil")
	}
	if found.UserID() != userID {
		t.Errorf("expected userID %s, got %s", userID, found.UserID())
	}
	if found.Token() != "valid-token-abc123" {
		t.Errorf("expected token valid-token-abc123, got %s", found.Token())
	}
	if found.UsedAt() != nil {
		t.Error("expected UsedAt to be nil on fresh token")
	}
}

func TestPasswordResetRepo_FindByToken_ExpiredReturnsNil(t *testing.T) {
	db := setupTestDB(t)
	userID := seedTestUser(t, db)

	repo := repository.NewGormPasswordResetRepository(db)
	ctx := context.Background()

	reset := identity.NewPasswordReset(userID, "expired-token-xyz", time.Now().UTC().Add(-time.Hour))
	if err := repo.Create(ctx, reset); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByToken(ctx, "expired-token-xyz")
	if found != nil {
		t.Error("expected nil for expired token, got non-nil")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected domain.ErrNotFound, got %v", err)
	}
}
