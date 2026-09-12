package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/domain"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/identity"
	"gorm.io/gorm"
)

// RunnerApplicationModel is the GORM model for the runner_applications table.
type RunnerApplicationModel struct {
	ID                      uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name                    string         `gorm:"type:text;not null"`
	Phone                   string         `gorm:"type:text;not null"`
	ICNumber                string         `gorm:"type:text;not null;column:ic_number"`
	VehicleType             string         `gorm:"type:text;not null;column:vehicle_type"`
	PlateNumber             string         `gorm:"type:text;not null;column:plate_number"`
	PetExperience           pq.StringArray `gorm:"type:text[];column:pet_experience"`
	ComfortableWithLivePets bool           `gorm:"column:comfortable_with_live_pets"`
	ConsentAcknowledged     bool           `gorm:"not null;column:consent_acknowledged"`
	Status                  string         `gorm:"type:text;not null;default:pending_review"`
	SubmittedAt             time.Time      `gorm:"not null;default:now();column:submitted_at"`
	ReviewedAt              *time.Time     `gorm:"column:reviewed_at"`
	ReviewerUserID          *uuid.UUID     `gorm:"type:uuid;column:reviewer_user_id"`
}

// TableName specifies the table name for GORM.
func (RunnerApplicationModel) TableName() string {
	return "runner_applications"
}

// fromDomainRunnerApplication converts a domain RunnerApplication to a RunnerApplicationModel.
func fromDomainRunnerApplication(a *identity.RunnerApplication) *RunnerApplicationModel {
	return &RunnerApplicationModel{
		ID:                      a.ID(),
		Name:                    a.Name(),
		Phone:                   a.Phone(),
		ICNumber:                a.ICNumber(),
		VehicleType:             a.VehicleType(),
		PlateNumber:             a.PlateNumber(),
		PetExperience:           pq.StringArray(a.PetExperience()),
		ComfortableWithLivePets: a.ComfortableWithLivePets(),
		ConsentAcknowledged:     a.ConsentAcknowledged(),
		Status:                  a.Status(),
		SubmittedAt:             a.SubmittedAt(),
		ReviewedAt:              a.ReviewedAt(),
		ReviewerUserID:          a.ReviewerUserID(),
	}
}

// GormRunnerApplicationRepository is a GORM-based implementation of RunnerApplicationRepository.
type GormRunnerApplicationRepository struct {
	db *gorm.DB
}

// NewGormRunnerApplicationRepository creates a new GormRunnerApplicationRepository.
func NewGormRunnerApplicationRepository(db *gorm.DB) *GormRunnerApplicationRepository {
	return &GormRunnerApplicationRepository{db: db}
}

// Insert persists a new runner application and returns the formatted display ID (KR-YYYY-NNNNN).
// The insert and annual count query run inside a single transaction so the display ID reflects
// the row just inserted. Under high concurrent load a small sequence gap or repeat is possible.
// Returns domain.NewAlreadyExistsError if ic_number already exists.
func (r *GormRunnerApplicationRepository) Insert(ctx context.Context, app *identity.RunnerApplication) (string, error) {
	model := fromDomainRunnerApplication(app)
	var displayID string

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(model).Error; err != nil {
			if isICNumberDuplicateError(err) {
				return domain.NewAlreadyExistsError("RunnerApplication", "ic_number", app.ICNumber())
			}
			return err
		}

		var count int64
		year := time.Now().UTC().Year()
		if err := tx.Model(&RunnerApplicationModel{}).
			Where("EXTRACT(YEAR FROM submitted_at) = ?", year).
			Count(&count).Error; err != nil {
			return err
		}

		displayID = fmt.Sprintf("KR-%d-%05d", year, count)
		return nil
	})
	if err != nil {
		return "", err
	}
	return displayID, nil
}

// isICNumberDuplicateError returns true if the error is a Postgres unique-constraint
// violation on the ic_number column.
func isICNumberDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// pq.Error carries the Postgres error code; 23505 = unique_violation.
	var pqErr *pq.Error
	if ok := isPqError(err, &pqErr); ok {
		return pqErr.Code == "23505" && strings.Contains(pqErr.Constraint, "ic_number")
	}
	// Fallback: string match on GORM's formatted error (works regardless of driver type).
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") && strings.Contains(msg, "ic_number")
}

// isPqError attempts to unwrap a pq.Error from err. Returns true if successful.
func isPqError(err error, target **pq.Error) bool {
	return errors.As(err, target)
}
