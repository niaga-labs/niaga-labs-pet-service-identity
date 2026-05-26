package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/domain"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/domain/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserModel is the GORM model for the users table.
type UserModel struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Email        string        `gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone        string        `gorm:"type:varchar(20)"`
	PasswordHash string        `gorm:"type:varchar(255);not null"`
	FullName     string        `gorm:"type:varchar(255);not null"`
	Role         auth.UserRole `gorm:"type:varchar(20);not null"`
	IsVerified   bool          `gorm:"default:false"`
	AvatarURL    string        `gorm:"type:text"`
	Version      int64         `gorm:"not null;default:1"`
	CreatedAt    time.Time     `gorm:"not null;default:now()"`
	UpdatedAt    time.Time     `gorm:"not null;default:now()"`
}

// TableName specifies the table name for GORM.
func (UserModel) TableName() string {
	return "users"
}

// UserRoleModel is a scoped role assignment. Global roles continue to live on
// users.role for backward compatibility; shop roles live here.
type UserRoleModel struct {
	UserID    uuid.UUID     `gorm:"type:uuid;primaryKey"`
	Role      auth.UserRole `gorm:"type:varchar(30);primaryKey"`
	ScopeType string        `gorm:"type:varchar(20);primaryKey"`
	ScopeID   uuid.UUID     `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time     `gorm:"not null;default:now()"`
}

// TableName specifies the table name for scoped roles.
func (UserRoleModel) TableName() string { return "user_roles" }

// toDomain converts a UserModel to a domain User.
func (m *UserModel) toDomain() *identity.User {
	return identity.ReconstructUser(
		m.ID,
		m.Email,
		m.Phone,
		m.PasswordHash,
		m.FullName,
		m.Role,
		m.IsVerified,
		m.AvatarURL,
		m.Version,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

// fromDomainUser converts a domain User to a UserModel.
func fromDomainUser(u *identity.User) *UserModel {
	return &UserModel{
		ID:           u.ID(),
		Email:        u.Email(),
		Phone:        u.Phone(),
		PasswordHash: u.PasswordHash(),
		FullName:     u.FullName(),
		Role:         u.Role(),
		IsVerified:   u.IsVerified(),
		AvatarURL:    u.AvatarURL(),
		Version:      u.Version(),
		CreatedAt:    u.CreatedAt(),
		UpdatedAt:    u.UpdatedAt(),
	}
}

// GormUserRepository is a GORM-based implementation of UserRepository.
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GormUserRepository.
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// FindByID retrieves a user by their unique ID.
func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// FindByEmail retrieves a user by their email address.
func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// Save persists a new user to the database.
func (r *GormUserRepository) Save(ctx context.Context, user *identity.User) error {
	model := fromDomainUser(user)
	return r.db.WithContext(ctx).Create(model).Error
}

// Update persists changes to an existing user with optimistic locking.
func (r *GormUserRepository) Update(ctx context.Context, user *identity.User) error {
	model := fromDomainUser(user)
	result := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ? AND version = ?", model.ID, model.Version-1).
		Updates(model)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.NewConflictError("user was modified by another transaction")
	}
	return nil
}

// ListAll returns a paginated list of all users.
func (r *GormUserRepository) ListAll(ctx context.Context, page, limit int) ([]*identity.User, int64, error) {
	var total int64
	r.db.WithContext(ctx).Model(&UserModel{}).Count(&total)

	var models []UserModel
	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*identity.User, len(models))
	for i := range models {
		users[i] = models[i].toDomain()
	}
	return users, total, nil
}

// UpdatePasswordHash directly updates a user's password hash by ID.
func (r *GormUserRepository) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	result := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash": passwordHash,
			"updated_at":    time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CountByRole returns user counts grouped by role.
func (r *GormUserRepository) CountByRole(ctx context.Context) (map[string]int64, error) {
	type roleCount struct {
		Role  string
		Count int64
	}
	var results []roleCount
	if err := r.db.WithContext(ctx).Model(&UserModel{}).
		Select("role, count(*) as count").
		Group("role").
		Find(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, rc := range results {
		counts[rc.Role] = rc.Count
	}
	return counts, nil
}

// GrantRole grants a scoped role idempotently.
func (r *GormUserRepository) GrantRole(ctx context.Context, userID uuid.UUID, role auth.UserRole, scopeType string, scopeID *uuid.UUID) error {
	id := uuid.Nil
	if scopeID != nil {
		id = *scopeID
	}
	row := UserRoleModel{
		UserID:    userID,
		Role:      role,
		ScopeType: scopeType,
		ScopeID:   id,
		CreatedAt: time.Now().UTC(),
	}
	return r.db.WithContext(ctx).FirstOrCreate(&row, UserRoleModel{
		UserID:    userID,
		Role:      role,
		ScopeType: scopeType,
		ScopeID:   id,
	}).Error
}

// RevokeRole removes a scoped role idempotently.
func (r *GormUserRepository) RevokeRole(ctx context.Context, userID uuid.UUID, role auth.UserRole, scopeType string, scopeID *uuid.UUID) error {
	id := uuid.Nil
	if scopeID != nil {
		id = *scopeID
	}
	return r.db.WithContext(ctx).Where("user_id = ? AND role = ? AND scope_type = ? AND scope_id = ?", userID, role, scopeType, id).Delete(&UserRoleModel{}).Error
}

// ListShopsForUser returns all shop-scoped roles for a user.
func (r *GormUserRepository) ListShopsForUser(ctx context.Context, userID uuid.UUID) ([]identity.ShopRole, error) {
	var rows []UserRoleModel
	if err := r.db.WithContext(ctx).Where("user_id = ? AND scope_type = ?", userID, "shop").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]identity.ShopRole, len(rows))
	for i, row := range rows {
		out[i] = identity.ShopRole{ShopID: row.ScopeID, Role: row.Role}
	}
	return out, nil
}
