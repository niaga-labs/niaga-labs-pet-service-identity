package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/domain"
	"github.com/niaga-labs/niaga-labs-pet-lib-proto/dto"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/handler"
	"go.uber.org/zap"
)

type fakeResetPasswordService struct {
	err error
}

func (f *fakeResetPasswordService) ResetPassword(_ context.Context, _ dto.ResetPasswordRequest) error {
	return f.err
}

func setupResetPasswordRouter(svc handler.ResetPasswordService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiV1 := r.Group("/api/v1")
	h := handler.NewResetPasswordHandler(svc, zap.NewNop())
	h.RegisterRoutes(apiV1)
	return r
}

func TestResetPassword_ExpiredToken_Returns400(t *testing.T) {
	svc := &fakeResetPasswordService{err: domain.NewValidationError("token expired")}
	r := setupResetPasswordRouter(svc)

	body := bytes.NewBufferString(`{"token":"some-expired-token","newPassword":"newpass123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for expired token, got %d", w.Code)
	}
}

func TestResetPassword_UsedToken_Returns400(t *testing.T) {
	svc := &fakeResetPasswordService{err: domain.NewValidationError("token already used")}
	r := setupResetPasswordRouter(svc)

	body := bytes.NewBufferString(`{"token":"some-used-token","newPassword":"newpass123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for used token, got %d", w.Code)
	}
}

func TestResetPassword_UnknownToken_Returns400(t *testing.T) {
	svc := &fakeResetPasswordService{err: domain.NewValidationError("invalid token")}
	r := setupResetPasswordRouter(svc)

	body := bytes.NewBufferString(`{"token":"nonexistent-token","newPassword":"newpass123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown token, got %d", w.Code)
	}
}

func TestResetPassword_WeakPassword_Returns400(t *testing.T) {
	svc := &fakeResetPasswordService{err: domain.NewValidationError("password must be at least 8 characters")}
	r := setupResetPasswordRouter(svc)

	body := bytes.NewBufferString(`{"token":"valid-token","newPassword":"short"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for weak password, got %d", w.Code)
	}
}

func TestResetPassword_MalformedRequest_Returns400(t *testing.T) {
	svc := &fakeResetPasswordService{err: nil}
	r := setupResetPasswordRouter(svc)

	body := bytes.NewBufferString(`{`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", w.Code)
	}
}
