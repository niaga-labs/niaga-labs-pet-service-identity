package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/niaga-labs/niaga-labs-pet-lib-proto/dto"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/handler"
	"go.uber.org/zap"
)

type fakeForgotPasswordService struct {
	err error
}

func (f *fakeForgotPasswordService) ForgotPassword(_ context.Context, _ dto.ForgotPasswordRequest) error {
	return f.err
}

func setupForgotPasswordRouter(svc handler.ForgotPasswordService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiV1 := r.Group("/api/v1")
	h := handler.NewForgotPasswordHandler(svc, zap.NewNop())
	h.RegisterRoutes(apiV1)
	return r
}

func TestForgotPassword_KnownEmail_Returns202(t *testing.T) {
	r := setupForgotPasswordRouter(&fakeForgotPasswordService{err: nil})

	body := bytes.NewBufferString(`{"email":"runner.test@kilat.my"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}
}

func TestForgotPassword_UnknownEmail_AlsoReturns202(t *testing.T) {
	r := setupForgotPasswordRouter(&fakeForgotPasswordService{err: nil})

	body := bytes.NewBufferString(`{"email":"noone@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202 for unknown email, got %d", w.Code)
	}
}

func TestForgotPassword_MalformedRequest_Returns400(t *testing.T) {
	r := setupForgotPasswordRouter(&fakeForgotPasswordService{err: nil})

	body := bytes.NewBufferString(`{`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", w.Code)
	}
}
