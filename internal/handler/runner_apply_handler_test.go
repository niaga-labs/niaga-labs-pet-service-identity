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

type fakeRunnerApplyService struct {
	returnID string
	err      error
}

func (f *fakeRunnerApplyService) Apply(_ context.Context, _ dto.RunnerApplicationRequest) (string, error) {
	return f.returnID, f.err
}

func setupRunnerApplyRouter(svc handler.RunnerApplyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiV1 := r.Group("/api/v1")
	h := handler.NewRunnerApplyHandler(svc, zap.NewNop())
	h.RegisterRoutes(apiV1)
	return r
}

func TestRunnerApply_MissingConsent_Returns400(t *testing.T) {
	svc := &fakeRunnerApplyService{
		err: domain.NewValidationError("consent must be acknowledged"),
	}
	r := setupRunnerApplyRouter(svc)

	body := bytes.NewBufferString(`{
		"name":"Test Runner",
		"phone":"0123456789",
		"icNumber":"950101012345",
		"vehicleType":"motorbike",
		"plateNumber":"ABC1234",
		"petExperience":["dogs"],
		"comfortableWithLivePets":true,
		"consentAcknowledged":false
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runners/apply", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing consent, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRunnerApply_InvalidVehicleType_Returns400(t *testing.T) {
	svc := &fakeRunnerApplyService{
		err: domain.NewValidationError("vehicleType must be one of: motorbike, car, bicycle"),
	}
	r := setupRunnerApplyRouter(svc)

	body := bytes.NewBufferString(`{
		"name":"Test Runner",
		"phone":"0123456789",
		"icNumber":"950101012345",
		"vehicleType":"boat",
		"plateNumber":"ABC1234",
		"petExperience":[],
		"comfortableWithLivePets":true,
		"consentAcknowledged":true
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runners/apply", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid vehicle type, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRunnerApply_MalformedRequest_Returns400(t *testing.T) {
	svc := &fakeRunnerApplyService{returnID: "KR-2026-00001"}
	r := setupRunnerApplyRouter(svc)

	body := bytes.NewBufferString(`{`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runners/apply", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", w.Code)
	}
}

func TestRunnerApply_ValidRequest_Returns201(t *testing.T) {
	svc := &fakeRunnerApplyService{returnID: "KR-2026-00001"}
	r := setupRunnerApplyRouter(svc)

	body := bytes.NewBufferString(`{
		"name":"Test Runner",
		"phone":"0123456789",
		"icNumber":"950101012345",
		"vehicleType":"motorbike",
		"plateNumber":"ABC1234",
		"petExperience":["dogs","cats"],
		"comfortableWithLivePets":true,
		"consentAcknowledged":true
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runners/apply", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}
