package identity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	DocumentKindIC         = "ic"
	DocumentKindLicense    = "license"
	DocumentKindSelfie     = "selfie"
	DocumentKindVehicleReg = "vehicle_reg"

	DocumentStatusPending  = "pending"
	DocumentStatusVerified = "verified"
	DocumentStatusRejected = "rejected"
)

// UserDocument stores one verification artifact uploaded by a user.
type UserDocument struct {
	id               uuid.UUID
	userID           uuid.UUID
	kind             string
	storageURL       string
	status           string
	reviewedByUserID *uuid.UUID
	reviewedAt       *time.Time
	createdAt        time.Time
}

// NewUserDocument creates a pending document.
func NewUserDocument(userID uuid.UUID, kind, storageURL string) (*UserDocument, error) {
	if !validDocumentKind(kind) {
		return nil, fmt.Errorf("document kind must be one of: ic, license, selfie, vehicle_reg")
	}
	if storageURL == "" {
		return nil, fmt.Errorf("storage URL is required")
	}
	return &UserDocument{
		id:         uuid.New(),
		userID:     userID,
		kind:       kind,
		storageURL: storageURL,
		status:     DocumentStatusPending,
		createdAt:  time.Now().UTC(),
	}, nil
}

// ReconstructUserDocument rebuilds a document from persistence data.
func ReconstructUserDocument(id, userID uuid.UUID, kind, storageURL, status string, reviewedBy *uuid.UUID, reviewedAt *time.Time, createdAt time.Time) *UserDocument {
	return &UserDocument{
		id:               id,
		userID:           userID,
		kind:             kind,
		storageURL:       storageURL,
		status:           status,
		reviewedByUserID: reviewedBy,
		reviewedAt:       reviewedAt,
		createdAt:        createdAt,
	}
}

func (d *UserDocument) ID() uuid.UUID                { return d.id }
func (d *UserDocument) UserID() uuid.UUID            { return d.userID }
func (d *UserDocument) Kind() string                 { return d.kind }
func (d *UserDocument) StorageURL() string           { return d.storageURL }
func (d *UserDocument) Status() string               { return d.status }
func (d *UserDocument) ReviewedByUserID() *uuid.UUID { return d.reviewedByUserID }
func (d *UserDocument) ReviewedAt() *time.Time       { return d.reviewedAt }
func (d *UserDocument) CreatedAt() time.Time         { return d.createdAt }

// Review records an admin review decision.
func (d *UserDocument) Review(reviewerID uuid.UUID, status string) error {
	if status != DocumentStatusVerified && status != DocumentStatusRejected {
		return fmt.Errorf("review status must be verified or rejected")
	}
	now := time.Now().UTC()
	d.status = status
	d.reviewedByUserID = &reviewerID
	d.reviewedAt = &now
	return nil
}

// IsValidDocumentKind reports whether kind is supported.
func IsValidDocumentKind(kind string) bool {
	return kind == DocumentKindIC ||
		kind == DocumentKindLicense ||
		kind == DocumentKindSelfie ||
		kind == DocumentKindVehicleReg
}

func validDocumentKind(kind string) bool {
	return IsValidDocumentKind(kind)
}
