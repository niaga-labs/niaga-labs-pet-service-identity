package referral

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// Referral represents a referral relationship between two users.
type Referral struct {
	id                uuid.UUID
	referrerID        uuid.UUID
	refereeID         uuid.UUID
	referralCode      string
	rewardAmountCents int64
	referrerCredited  bool
	refereeCredited   bool
	createdAt         time.Time
}

// NewReferral creates a new referral record.
func NewReferral(referrerID, refereeID uuid.UUID, referralCode string, rewardAmountCents int64) *Referral {
	return &Referral{
		id:                uuid.New(),
		referrerID:        referrerID,
		refereeID:         refereeID,
		referralCode:      referralCode,
		rewardAmountCents: rewardAmountCents,
		referrerCredited:  false,
		refereeCredited:   false,
		createdAt:         time.Now().UTC(),
	}
}

// Reconstruct rebuilds a Referral from persistence.
func Reconstruct(id, referrerID, refereeID uuid.UUID, referralCode string, rewardAmountCents int64, referrerCredited, refereeCredited bool, createdAt time.Time) *Referral {
	return &Referral{
		id: id, referrerID: referrerID, refereeID: refereeID,
		referralCode: referralCode, rewardAmountCents: rewardAmountCents,
		referrerCredited: referrerCredited, refereeCredited: refereeCredited,
		createdAt: createdAt,
	}
}

// CreditReferrer marks the referrer as credited.
func (r *Referral) CreditReferrer() { r.referrerCredited = true }

// CreditReferee marks the referee as credited.
func (r *Referral) CreditReferee() { r.refereeCredited = true }

// Getters.
func (r *Referral) ID() uuid.UUID            { return r.id }
func (r *Referral) ReferrerID() uuid.UUID    { return r.referrerID }
func (r *Referral) RefereeID() uuid.UUID     { return r.refereeID }
func (r *Referral) ReferralCode() string     { return r.referralCode }
func (r *Referral) RewardAmountCents() int64 { return r.rewardAmountCents }
func (r *Referral) ReferrerCredited() bool   { return r.referrerCredited }
func (r *Referral) RefereeCredited() bool    { return r.refereeCredited }
func (r *Referral) CreatedAt() time.Time     { return r.createdAt }

// GenerateReferralCode creates a unique referral code.
func GenerateReferralCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "REF-" + hex.EncodeToString(b), nil
}
