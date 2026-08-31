package identity

import (
	"time"

	"github.com/google/uuid"
)

// RunnerApplication represents a runner's application to join the Kilat platform.
// Fields are private; access via getters only.
type RunnerApplication struct {
	id                      uuid.UUID
	name                    string
	phone                   string
	icNumber                string
	vehicleType             string
	plateNumber             string
	petExperience           []string
	comfortableWithLivePets bool
	consentAcknowledged     bool
	status                  string
	submittedAt             time.Time
	reviewedAt              *time.Time
	reviewerUserID          *uuid.UUID
}

// NewRunnerApplication creates a new RunnerApplication with status pending_review.
func NewRunnerApplication(
	name, phone, icNumber, vehicleType, plateNumber string,
	petExperience []string,
	comfortableWithLivePets, consentAcknowledged bool,
) *RunnerApplication {
	return &RunnerApplication{
		id:                      uuid.New(),
		name:                    name,
		phone:                   phone,
		icNumber:                icNumber,
		vehicleType:             vehicleType,
		plateNumber:             plateNumber,
		petExperience:           petExperience,
		comfortableWithLivePets: comfortableWithLivePets,
		consentAcknowledged:     consentAcknowledged,
		status:                  "pending_review",
		submittedAt:             time.Now().UTC(),
		reviewedAt:              nil,
		reviewerUserID:          nil,
	}
}

// --- Getters ---

// ID returns the application's unique UUID.
func (r *RunnerApplication) ID() uuid.UUID { return r.id }

// Name returns the applicant's full name.
func (r *RunnerApplication) Name() string { return r.name }

// Phone returns the applicant's phone number.
func (r *RunnerApplication) Phone() string { return r.phone }

// ICNumber returns the applicant's IC/NRIC number.
func (r *RunnerApplication) ICNumber() string { return r.icNumber }

// VehicleType returns the type of vehicle (motorbike, car, bicycle).
func (r *RunnerApplication) VehicleType() string { return r.vehicleType }

// PlateNumber returns the vehicle plate number.
func (r *RunnerApplication) PlateNumber() string { return r.plateNumber }

// PetExperience returns the list of pet types the applicant has experience with.
func (r *RunnerApplication) PetExperience() []string { return r.petExperience }

// ComfortableWithLivePets returns whether the applicant is comfortable with live pets.
func (r *RunnerApplication) ComfortableWithLivePets() bool { return r.comfortableWithLivePets }

// ConsentAcknowledged returns whether the applicant has acknowledged the consent.
func (r *RunnerApplication) ConsentAcknowledged() bool { return r.consentAcknowledged }

// Status returns the current review status.
func (r *RunnerApplication) Status() string { return r.status }

// SubmittedAt returns when the application was submitted.
func (r *RunnerApplication) SubmittedAt() time.Time { return r.submittedAt }

// ReviewedAt returns when the application was reviewed, or nil if not yet reviewed.
func (r *RunnerApplication) ReviewedAt() *time.Time { return r.reviewedAt }

// ReviewerUserID returns the ID of the reviewer, or nil if not yet reviewed.
func (r *RunnerApplication) ReviewerUserID() *uuid.UUID { return r.reviewerUserID }
