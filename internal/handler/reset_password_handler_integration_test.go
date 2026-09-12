//go:build integration

package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/auth"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/application"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/identity"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/handler"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newTestJWTManager() *auth.JWTManager {
	return auth.NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
}

type fakeResetNotifier struct{}

func (f *fakeResetNotifier) SendPasswordResetEmail(_ context.Context, _, _ string) error {
	return nil
}

func setupIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5435 user=kilat password=kilat_secret dbname=kilat_identity sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to integration test db: %v", err)
	}
	return db
}

func seedIntegrationUser(t *testing.T, db *gorm.DB) (uuid.UUID, string) {
	t.Helper()
	originalPassword := "originalpass99"
	hash, err := bcrypt.GenerateFromPassword([]byte(originalPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt failed: %v", err)
	}
	userID := uuid.New()
	email := uuid.NewString() + "@reset-integration-test.local"
	model := repository.UserModel{
		ID:           userID,
		Email:        email,
		Phone:        "",
		PasswordHash: string(hash),
		FullName:     "Integration Reset User",
		Role:         "runner",
	}
	if err := db.Create(&model).Error; err != nil {
		t.Fatalf("seed integration user failed: %v", err)
	}
	return userID, originalPassword
}

func seedIntegrationToken(t *testing.T, db *gorm.DB, userID uuid.UUID) string {
	t.Helper()
	token := "integration-reset-token-" + uuid.NewString()
	reset := identity.NewPasswordReset(userID, token, time.Now().UTC().Add(time.Hour))
	repo := repository.NewGormPasswordResetRepository(db)
	if err := repo.Create(context.Background(), reset); err != nil {
		t.Fatalf("seed password reset token failed: %v", err)
	}
	return token
}

func TestResetPassword_ValidToken_Returns200_UpdatesHash(t *testing.T) {
	db := setupIntegrationDB(t)

	userID, _ := seedIntegrationUser(t, db)
	t.Cleanup(func() {
		db.Where("id = ?", userID).Delete(&repository.UserModel{})
	})

	token := seedIntegrationToken(t, db, userID)
	t.Cleanup(func() {
		db.Where("user_id = ?", userID).Delete(&repository.PasswordResetModel{})
	})

	userRepo := repository.NewGormUserRepository(db)
	tokenRepo := repository.NewGormTokenRepository(db)
	passwordResetRepo := repository.NewGormPasswordResetRepository(db)
	notifier := &fakeResetNotifier{}
	logger := zap.NewNop()

	jwtManager := newTestJWTManager()
	authService := application.NewAuthService(userRepo, tokenRepo, passwordResetRepo, notifier, jwtManager, logger)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiV1 := r.Group("/api/v1")
	h := handler.NewResetPasswordHandler(authService, logger)
	h.RegisterRoutes(apiV1)

	newPassword := "newpass1234"
	body := bytes.NewBufferString(`{"token":"` + token + `","newPassword":"` + newPassword + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	// Verify password hash changed in DB
	var userModel repository.UserModel
	if err := db.Where("id = ?", userID).First(&userModel).Error; err != nil {
		t.Fatalf("failed to read user after reset: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(newPassword)); err != nil {
		t.Errorf("new password hash does not match expected new password: %v", err)
	}

	// Verify token marked as used
	var resetModel repository.PasswordResetModel
	if err := db.Where("token = ?", token).First(&resetModel).Error; err != nil {
		t.Fatalf("failed to read password_reset after reset: %v", err)
	}
	if resetModel.UsedAt == nil {
		t.Error("expected used_at to be set after reset, but it is nil")
	}
}
