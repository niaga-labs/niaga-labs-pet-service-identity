package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	referralDomain "github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/referral"
	"gorm.io/gorm"
)

// ReferralModel is the GORM model for the referrals table.
type ReferralModel struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReferrerID        uuid.UUID `gorm:"type:uuid;not null;index"`
	RefereeID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	ReferralCode      string    `gorm:"type:varchar(50);not null"`
	RewardAmountCents int64     `gorm:"default:0"`
	ReferrerCredited  bool      `gorm:"default:false"`
	RefereeCredited   bool      `gorm:"default:false"`
	CreatedAt         time.Time `gorm:"not null"`
}

// TableName sets the table name.
func (ReferralModel) TableName() string { return "referrals" }

// UserReferralCodeModel stores each user's unique referral code.
type UserReferralCodeModel struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Code   string    `gorm:"type:varchar(50);uniqueIndex;not null"`
}

// TableName sets the table name.
func (UserReferralCodeModel) TableName() string { return "user_referral_codes" }

// GormReferralRepository implements ReferralRepository using GORM.
type GormReferralRepository struct {
	db *gorm.DB
}

// NewGormReferralRepository creates a new GormReferralRepository.
func NewGormReferralRepository(db *gorm.DB) *GormReferralRepository {
	return &GormReferralRepository{db: db}
}

// Save persists a new referral.
func (r *GormReferralRepository) Save(ctx context.Context, ref *referralDomain.Referral) error {
	model := toReferralModel(ref)
	return r.db.WithContext(ctx).Create(&model).Error
}

// Update updates a referral.
func (r *GormReferralRepository) Update(ctx context.Context, ref *referralDomain.Referral) error {
	model := toReferralModel(ref)
	return r.db.WithContext(ctx).Save(&model).Error
}

// FindByReferrerID returns all referrals made by a user.
func (r *GormReferralRepository) FindByReferrerID(ctx context.Context, referrerID uuid.UUID) ([]*referralDomain.Referral, error) {
	var models []ReferralModel
	if err := r.db.WithContext(ctx).Where("referrer_id = ?", referrerID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	refs := make([]*referralDomain.Referral, len(models))
	for i, m := range models {
		refs[i] = toReferralDomain(&m)
	}
	return refs, nil
}

// FindByReferralCode returns the referral that used a specific code.
func (r *GormReferralRepository) FindByReferralCode(ctx context.Context, code string) (*referralDomain.Referral, error) {
	var model ReferralModel
	if err := r.db.WithContext(ctx).Where("referral_code = ?", code).First(&model).Error; err != nil {
		return nil, err
	}
	return toReferralDomain(&model), nil
}

// FindByRefereeID returns the referral for a referee.
func (r *GormReferralRepository) FindByRefereeID(ctx context.Context, refereeID uuid.UUID) (*referralDomain.Referral, error) {
	var model ReferralModel
	if err := r.db.WithContext(ctx).Where("referee_id = ?", refereeID).First(&model).Error; err != nil {
		return nil, err
	}
	return toReferralDomain(&model), nil
}

// CountByReferrerID returns the number of successful referrals.
func (r *GormReferralRepository) CountByReferrerID(ctx context.Context, referrerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ReferralModel{}).Where("referrer_id = ?", referrerID).Count(&count).Error
	return count, err
}

// SaveUserReferralCode saves a user's unique referral code.
func (r *GormReferralRepository) SaveUserReferralCode(ctx context.Context, userID uuid.UUID, code string) error {
	model := UserReferralCodeModel{
		ID:     uuid.New(),
		UserID: userID,
		Code:   code,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

// GetUserReferralCode returns a user's referral code.
func (r *GormReferralRepository) GetUserReferralCode(ctx context.Context, userID uuid.UUID) (string, error) {
	var model UserReferralCodeModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		return "", err
	}
	return model.Code, nil
}

// FindUserIDByReferralCode finds the user who owns a referral code.
func (r *GormReferralRepository) FindUserIDByReferralCode(ctx context.Context, code string) (uuid.UUID, error) {
	var model UserReferralCodeModel
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&model).Error; err != nil {
		return uuid.Nil, err
	}
	return model.UserID, nil
}

func toReferralModel(r *referralDomain.Referral) ReferralModel {
	return ReferralModel{
		ID:                r.ID(),
		ReferrerID:        r.ReferrerID(),
		RefereeID:         r.RefereeID(),
		ReferralCode:      r.ReferralCode(),
		RewardAmountCents: r.RewardAmountCents(),
		ReferrerCredited:  r.ReferrerCredited(),
		RefereeCredited:   r.RefereeCredited(),
		CreatedAt:         r.CreatedAt(),
	}
}

func toReferralDomain(m *ReferralModel) *referralDomain.Referral {
	return referralDomain.Reconstruct(
		m.ID, m.ReferrerID, m.RefereeID,
		m.ReferralCode, m.RewardAmountCents,
		m.ReferrerCredited, m.RefereeCredited,
		m.CreatedAt,
	)
}
