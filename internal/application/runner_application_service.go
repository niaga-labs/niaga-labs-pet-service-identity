package application

import (
	"context"

	"github.com/niaga-labs/niaga-labs-pet-lib-common/domain"
	"github.com/niaga-labs/niaga-labs-pet-lib-proto/dto"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/identity"
	"go.uber.org/zap"
)

// RunnerApplicationService handles the runner application use case.
type RunnerApplicationService struct {
	repo   identity.RunnerApplicationRepository
	logger *zap.Logger
}

// NewRunnerApplicationService creates a new RunnerApplicationService.
func NewRunnerApplicationService(repo identity.RunnerApplicationRepository, logger *zap.Logger) *RunnerApplicationService {
	return &RunnerApplicationService{
		repo:   repo,
		logger: logger,
	}
}

// Apply validates the request, constructs the domain entity, and persists it.
// Returns the formatted display ID (e.g. KR-2026-00001) on success.
func (s *RunnerApplicationService) Apply(ctx context.Context, req dto.RunnerApplicationRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", domain.NewValidationError(err.Error())
	}

	app := identity.NewRunnerApplication(
		req.Name,
		req.Phone,
		req.ICNumber,
		req.VehicleType,
		req.PlateNumber,
		req.PetExperience,
		req.ComfortableWithLivePets,
		req.ConsentAcknowledged,
	)

	displayID, err := s.repo.Insert(ctx, app)
	if err != nil {
		s.logger.Error("runner application insert failed", zap.Error(err))
		return "", err
	}

	s.logger.Info("runner application submitted",
		zap.String("display_id", displayID),
	)
	return displayID, nil
}
