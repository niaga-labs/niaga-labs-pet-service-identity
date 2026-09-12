package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	referralDomain "github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/referral"
	"go.uber.org/zap"
)

const defaultRewardCents = 500 // RM 5.00 reward per referral

// ReferralStatsDTO is the API response for referral statistics.
type ReferralStatsDTO struct {
	ReferralCode   string         `json:"referral_code"`
	TotalReferrals int64          `json:"total_referrals"`
	Referrals      []*ReferralDTO `json:"referrals"`
}

// ReferralDTO is the API response for a single referral.
type ReferralDTO struct {
	ID                uuid.UUID `json:"id"`
	RefereeID         uuid.UUID `json:"referee_id"`
	RewardAmountCents int64     `json:"reward_amount_cents"`
	ReferrerCredited  bool      `json:"referrer_credited"`
	RefereeCredited   bool      `json:"referee_credited"`
	CreatedAt         time.Time `json:"created_at"`
}

// ReferralService handles referral use cases.
type ReferralService struct {
	repo   referralDomain.ReferralRepository
	logger *zap.Logger
}

// NewReferralService creates a new ReferralService.
func NewReferralService(repo referralDomain.ReferralRepository, logger *zap.Logger) *ReferralService {
	return &ReferralService{repo: repo, logger: logger}
}

// GetOrCreateReferralCode returns the user's referral code, creating one if needed.
func (s *ReferralService) GetOrCreateReferralCode(ctx context.Context, userID uuid.UUID) (string, error) {
	code, err := s.repo.GetUserReferralCode(ctx, userID)
	if err == nil && code != "" {
		return code, nil
	}

	// Generate a new code
	code, err = referralDomain.GenerateReferralCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate referral code: %w", err)
	}

	if err := s.repo.SaveUserReferralCode(ctx, userID, code); err != nil {
		return "", fmt.Errorf("failed to save referral code: %w", err)
	}

	s.logger.Info("referral code created", zap.String("user_id", userID.String()), zap.String("code", code))
	return code, nil
}

// ProcessReferral creates a referral record when a new user registers with a referral code.
func (s *ReferralService) ProcessReferral(ctx context.Context, refereeID uuid.UUID, referralCode string) error {
	if referralCode == "" {
		return nil
	}

	// Find who owns this referral code
	ownerCode, err := s.findCodeOwner(ctx, referralCode)
	if err != nil {
		s.logger.Warn("invalid referral code", zap.String("code", referralCode))
		return nil // Don't block registration for invalid codes
	}

	ref := referralDomain.NewReferral(ownerCode, refereeID, referralCode, defaultRewardCents)
	if err := s.repo.Save(ctx, ref); err != nil {
		s.logger.Error("failed to save referral", zap.Error(err))
		return nil // Don't block registration
	}

	s.logger.Info("referral processed",
		zap.String("referrer_id", ownerCode.String()),
		zap.String("referee_id", refereeID.String()),
	)
	return nil
}

// GetMyReferrals returns the user's referral stats and list.
func (s *ReferralService) GetMyReferrals(ctx context.Context, userID uuid.UUID) (*ReferralStatsDTO, error) {
	code, _ := s.GetOrCreateReferralCode(ctx, userID)

	referrals, err := s.repo.FindByReferrerID(ctx, userID)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountByReferrerID(ctx, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*ReferralDTO, len(referrals))
	for i, r := range referrals {
		dtos[i] = &ReferralDTO{
			ID:                r.ID(),
			RefereeID:         r.RefereeID(),
			RewardAmountCents: r.RewardAmountCents(),
			ReferrerCredited:  r.ReferrerCredited(),
			RefereeCredited:   r.RefereeCredited(),
			CreatedAt:         r.CreatedAt(),
		}
	}

	return &ReferralStatsDTO{
		ReferralCode:   code,
		TotalReferrals: total,
		Referrals:      dtos,
	}, nil
}

// findCodeOwner looks up the user who owns a referral code.
func (s *ReferralService) findCodeOwner(ctx context.Context, code string) (uuid.UUID, error) {
	return s.repo.FindUserIDByReferralCode(ctx, code)
}
