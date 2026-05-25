package application_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/domain"
	"github.com/Kilat-Pet-Delivery/lib-common/storage"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/application"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/domain/identity"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type memoryUserRepo struct {
	users map[uuid.UUID]*identity.User
}

func newMemoryUserRepo(users ...*identity.User) *memoryUserRepo {
	repo := &memoryUserRepo{users: map[uuid.UUID]*identity.User{}}
	for _, user := range users {
		repo.users[user.ID()] = user
	}
	return repo
}

func (r *memoryUserRepo) FindByID(_ context.Context, id uuid.UUID) (*identity.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (r *memoryUserRepo) FindByEmail(_ context.Context, email string) (*identity.User, error) {
	for _, user := range r.users {
		if user.Email() == email {
			return user, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *memoryUserRepo) Save(_ context.Context, user *identity.User) error {
	r.users[user.ID()] = user
	return nil
}

func (r *memoryUserRepo) Update(_ context.Context, user *identity.User) error {
	r.users[user.ID()] = user
	return nil
}

func (r *memoryUserRepo) ListAll(_ context.Context, _ int, _ int) ([]*identity.User, int64, error) {
	users := make([]*identity.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, int64(len(users)), nil
}

func (r *memoryUserRepo) ListByRole(_ context.Context, role auth.UserRole, limit int) ([]*identity.User, error) {
	if limit < 1 {
		limit = 50
	}
	users := make([]*identity.User, 0, limit)
	for _, user := range r.users {
		if user.Role() == role {
			users = append(users, user)
			if len(users) == limit {
				break
			}
		}
	}
	return users, nil
}

func (r *memoryUserRepo) CountByRole(_ context.Context) (map[string]int64, error) {
	return map[string]int64{}, nil
}

func (r *memoryUserRepo) UpdatePasswordHash(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

type memorySettingsRepo struct {
	settings map[uuid.UUID]*identity.UserSettings
}

func newMemorySettingsRepo() *memorySettingsRepo {
	return &memorySettingsRepo{settings: map[uuid.UUID]*identity.UserSettings{}}
}

func (r *memorySettingsRepo) FindByUserID(_ context.Context, userID uuid.UUID) (*identity.UserSettings, error) {
	settings, ok := r.settings[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return settings, nil
}

func (r *memorySettingsRepo) Upsert(_ context.Context, settings *identity.UserSettings) error {
	r.settings[settings.UserID()] = settings
	return nil
}

type memoryDocumentRepo struct {
	documents map[uuid.UUID]*identity.UserDocument
}

func newMemoryDocumentRepo() *memoryDocumentRepo {
	return &memoryDocumentRepo{documents: map[uuid.UUID]*identity.UserDocument{}}
}

func (r *memoryDocumentRepo) Save(_ context.Context, document *identity.UserDocument) error {
	r.documents[document.ID()] = document
	return nil
}

func (r *memoryDocumentRepo) FindByID(_ context.Context, id uuid.UUID) (*identity.UserDocument, error) {
	document, ok := r.documents[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return document, nil
}

func (r *memoryDocumentRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]*identity.UserDocument, error) {
	documents := []*identity.UserDocument{}
	for _, document := range r.documents {
		if document.UserID() == userID {
			documents = append(documents, document)
		}
	}
	return documents, nil
}

func (r *memoryDocumentRepo) Update(_ context.Context, document *identity.UserDocument) error {
	r.documents[document.ID()] = document
	return nil
}

func TestProfileServiceUpdateProfilePersistsChanges(t *testing.T) {
	ctx := context.Background()
	user := mustUser(t, "owner@example.com", auth.RoleOwner)
	userRepo := newMemoryUserRepo(user)
	service := application.NewProfileService(userRepo, newMemorySettingsRepo(), storage.NewFakeStorage(), zap.NewNop())

	result, err := service.UpdateProfile(ctx, user.ID(), application.UpdateProfileRequest{
		FullName: "Updated Owner",
		Phone:    "+60123456789",
	})
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}
	if result.FullName != "Updated Owner" || result.Phone != "+60123456789" {
		t.Fatalf("expected updated profile, got %+v", result)
	}

	persisted, _ := userRepo.FindByID(ctx, user.ID())
	if persisted.FullName() != "Updated Owner" || persisted.Phone() != "+60123456789" {
		t.Fatalf("expected persisted changes, got %s / %s", persisted.FullName(), persisted.Phone())
	}
}

func TestProfileServiceUploadPhotoReplacesOldStorageKey(t *testing.T) {
	ctx := context.Background()
	user := mustUser(t, "runner@example.com", auth.RoleRunner)
	userRepo := newMemoryUserRepo(user)
	fakeStorage := storage.NewFakeStorage()
	service := application.NewProfileService(userRepo, newMemorySettingsRepo(), fakeStorage, zap.NewNop())

	first, err := service.UploadProfilePhoto(ctx, user.ID(), "first.jpg", bytes.NewBufferString("first"))
	if err != nil {
		t.Fatalf("first upload returned error: %v", err)
	}
	second, err := service.UploadProfilePhoto(ctx, user.ID(), "second.jpg", bytes.NewBufferString("second"))
	if err != nil {
		t.Fatalf("second upload returned error: %v", err)
	}
	if first.AvatarURL == second.AvatarURL {
		t.Fatalf("expected replacement URL to change")
	}
	if _, err := fakeStorage.Download(ctx, first.AvatarURL); !errors.Is(err, storage.ErrObjectNotFound) {
		t.Fatalf("expected old storage key to be deleted, got %v", err)
	}
}

func TestProfileServiceReturnsDefaultSettingsBeforeFirstWrite(t *testing.T) {
	user := mustUser(t, "settings@example.com", auth.RoleOwner)
	service := application.NewProfileService(newMemoryUserRepo(user), newMemorySettingsRepo(), storage.NewFakeStorage(), zap.NewNop())

	result, err := service.GetSettings(context.Background(), user.ID())
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if result.Language != identity.LanguageEnglish || result.Theme != identity.ThemeSystem {
		t.Fatalf("expected default settings, got %+v", result)
	}
	if !result.NotificationsEnabled["push"] {
		t.Fatalf("expected push notifications enabled by default")
	}
}

func TestDocumentServiceUploadDocumentStartsPending(t *testing.T) {
	userID := uuid.New()
	documentRepo := newMemoryDocumentRepo()
	service := application.NewDocumentService(documentRepo, storage.NewFakeStorage(), zap.NewNop())

	result, err := service.UploadDocument(context.Background(), userID, identity.DocumentKindLicense, "license.png", bytes.NewBufferString("license"))
	if err != nil {
		t.Fatalf("UploadDocument returned error: %v", err)
	}
	if result.Status != identity.DocumentStatusPending {
		t.Fatalf("expected pending status, got %s", result.Status)
	}
}

func TestDocumentServiceReviewRecordsReviewerAndTimestamp(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	reviewerID := uuid.New()
	documentRepo := newMemoryDocumentRepo()
	service := application.NewDocumentService(documentRepo, storage.NewFakeStorage(), zap.NewNop())
	uploaded, err := service.UploadDocument(ctx, userID, identity.DocumentKindIC, "ic.png", bytes.NewBufferString("ic"))
	if err != nil {
		t.Fatalf("UploadDocument returned error: %v", err)
	}

	result, err := service.ReviewDocument(ctx, uploaded.ID, reviewerID, identity.DocumentStatusVerified)
	if err != nil {
		t.Fatalf("ReviewDocument returned error: %v", err)
	}
	if result.ReviewedByUserID == nil || *result.ReviewedByUserID != reviewerID {
		t.Fatalf("expected reviewer %s, got %+v", reviewerID, result.ReviewedByUserID)
	}
	if result.ReviewedAt == nil {
		t.Fatalf("expected review timestamp")
	}
}

func TestProfileServiceAgentsReturnsOnlySupportAgentRole(t *testing.T) {
	agent := mustUser(t, "agent@example.com", application.RoleSupportAgent)
	runner := mustUser(t, "runner2@example.com", auth.RoleRunner)
	service := application.NewProfileService(newMemoryUserRepo(agent, runner), newMemorySettingsRepo(), storage.NewFakeStorage(), zap.NewNop())

	result, err := service.ListAvailableAgents(context.Background(), 50)
	if err != nil {
		t.Fatalf("ListAvailableAgents returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected one support agent, got %d", len(result))
	}
	if result[0].ID != agent.ID() || result[0].Role != string(application.RoleSupportAgent) {
		t.Fatalf("expected support agent, got %+v", result[0])
	}
}

func mustUser(t *testing.T, email string, role auth.UserRole) *identity.User {
	t.Helper()
	user, err := identity.NewUser(email, "+60000000000", "Test User", "password-hash", role)
	if err != nil {
		t.Fatalf("NewUser returned error: %v", err)
	}
	user.Verify()
	return user
}
