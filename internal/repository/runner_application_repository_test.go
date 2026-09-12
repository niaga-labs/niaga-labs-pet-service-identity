//go:build integration

package repository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/domain"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/application"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/identity"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/handler"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/repository"
	"go.uber.org/zap"
)

// setupRunnerApplyChain builds the full handler → service → repo → DB chain.
func setupRunnerApplyChain(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	db := setupTestDB(t)

	repo := repository.NewGormRunnerApplicationRepository(db)
	svc := application.NewRunnerApplicationService(repo, zap.NewNop())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiV1 := r.Group("/api/v1")
	h := handler.NewRunnerApplyHandler(svc, zap.NewNop())
	h.RegisterRoutes(apiV1)

	cleanup := func() {
		db.Exec("TRUNCATE TABLE runner_applications RESTART IDENTITY CASCADE")
	}
	return r, cleanup
}

// uniqueIC generates a unique IC number per test run.
func uniqueIC() string {
	return fmt.Sprintf("IC-%s", uuid.NewString()[:13])
}

func TestRunnerApply_ValidRequest_Returns201WithApplicationID(t *testing.T) {
	r, cleanup := setupRunnerApplyChain(t)
	t.Cleanup(cleanup)

	db := setupTestDB(t)
	icNumber := uniqueIC()

	body := bytes.NewBufferString(fmt.Sprintf(`{
		"name":"Test Runner",
		"phone":"0123456789",
		"icNumber":%q,
		"vehicleType":"motorbike",
		"plateNumber":"ABC1234",
		"petExperience":["dogs","cats"],
		"comfortableWithLivePets":true,
		"consentAcknowledged":true
	}`, icNumber))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/runners/apply", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}

	// Parse response and verify applicationId format.
	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			ApplicationID string `json:"applicationId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success=true, got false")
	}
	matched, _ := regexp.MatchString(`^KR-\d{4}-\d{5}$`, resp.Data.ApplicationID)
	if !matched {
		t.Errorf("applicationId %q does not match KR-YYYY-NNNNN format", resp.Data.ApplicationID)
	}

	// Verify DB row exists with correct status.
	var model repository.RunnerApplicationModel
	if err := db.Where("ic_number = ?", icNumber).First(&model).Error; err != nil {
		t.Fatalf("expected DB row for ic_number %q, got error: %v", icNumber, err)
	}
	if model.Status != "pending_review" {
		t.Errorf("expected status 'pending_review', got %q", model.Status)
	}
}

func TestRunnerApply_DuplicateICNumber_Returns409(t *testing.T) {
	r, cleanup := setupRunnerApplyChain(t)
	t.Cleanup(cleanup)

	icNumber := uniqueIC()

	postApply := func() *httptest.ResponseRecorder {
		body := bytes.NewBufferString(fmt.Sprintf(`{
			"name":"Test Runner",
			"phone":"0123456789",
			"icNumber":%q,
			"vehicleType":"car",
			"plateNumber":"XYZ9999",
			"petExperience":[],
			"comfortableWithLivePets":false,
			"consentAcknowledged":true
		}`, icNumber))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/runners/apply", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	// First application — must succeed.
	w1 := postApply()
	if w1.Code != http.StatusCreated {
		t.Fatalf("expected 201 on first apply, got %d — body: %s", w1.Code, w1.Body.String())
	}

	// Second application with same IC — must return 409.
	w2 := postApply()
	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 on duplicate IC, got %d — body: %s", w2.Code, w2.Body.String())
	}
}

// TestRunnerApplicationRepo_Insert_AlreadyExistsError tests the repo directly.
func TestRunnerApplicationRepo_Insert_AlreadyExistsError(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE runner_applications RESTART IDENTITY CASCADE")
	})

	repo := repository.NewGormRunnerApplicationRepository(db)
	ctx := context.Background()

	icNumber := uniqueIC()
	app1 := identity.NewRunnerApplication(
		"Runner One", "0111111111", icNumber,
		"motorbike", "AAA1111",
		[]string{"dogs"}, true, true,
	)

	if _, err := repo.Insert(ctx, app1); err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	app2 := identity.NewRunnerApplication(
		"Runner Two", "0222222222", icNumber,
		"car", "BBB2222",
		[]string{}, false, true,
	)
	_, err := repo.Insert(ctx, app2)
	if err == nil {
		t.Fatal("expected error on duplicate IC, got nil")
	}
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected domain.ErrAlreadyExists, got %T: %v", err, err)
	}
}
