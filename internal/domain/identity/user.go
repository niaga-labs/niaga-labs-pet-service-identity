package identity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/auth"
)

// User is the aggregate root representing a system user.
type User struct {
	id           uuid.UUID
	email        Email
	phone        Phone
	passwordHash string
	fullName     string
	role         auth.UserRole
	isVerified   bool
	avatarURL    string
	version      int64
	createdAt    time.Time
	updatedAt    time.Time
}

// NewUser creates a new User with validated fields.
func NewUser(email, phone, fullName, passwordHash string, role auth.UserRole) (*User, error) {
	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}
	phoneVO, err := NewPhone(phone)
	if err != nil {
		return nil, err
	}
	if fullName == "" {
		return nil, fmt.Errorf("full name is required")
	}
	if passwordHash == "" {
		return nil, fmt.Errorf("password hash is required")
	}

	now := time.Now().UTC()
	return &User{
		id:           uuid.New(),
		email:        emailVO,
		phone:        phoneVO,
		passwordHash: passwordHash,
		fullName:     fullName,
		role:         role,
		isVerified:   false,
		avatarURL:    "",
		version:      1,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ReconstructUser rebuilds a User from persistence data (no validation).
func ReconstructUser(
	id uuid.UUID,
	email, phone, passwordHash, fullName string,
	role auth.UserRole,
	isVerified bool,
	avatarURL string,
	version int64,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		id:           id,
		email:        Email{value: email},
		phone:        Phone{value: phone},
		passwordHash: passwordHash,
		fullName:     fullName,
		role:         role,
		isVerified:   isVerified,
		avatarURL:    avatarURL,
		version:      version,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// --- Getters ---

// ID returns the user's unique identifier.
func (u *User) ID() uuid.UUID { return u.id }

// Email returns the user's email address as a string.
func (u *User) Email() string { return u.email.String() }

// EmailVO returns the user's email as a value object.
func (u *User) EmailVO() Email { return u.email }

// Phone returns the user's phone number as a string.
func (u *User) Phone() string { return u.phone.String() }

// PhoneVO returns the user's phone as a value object.
func (u *User) PhoneVO() Phone { return u.phone }

// PasswordHash returns the user's hashed password.
func (u *User) PasswordHash() string { return u.passwordHash }

// FullName returns the user's full name.
func (u *User) FullName() string { return u.fullName }

// Role returns the user's role.
func (u *User) Role() auth.UserRole { return u.role }

// IsVerified returns whether the user has been verified.
func (u *User) IsVerified() bool { return u.isVerified }

// AvatarURL returns the user's avatar URL.
func (u *User) AvatarURL() string { return u.avatarURL }

// Version returns the entity version for optimistic locking.
func (u *User) Version() int64 { return u.version }

// CreatedAt returns the creation timestamp.
func (u *User) CreatedAt() time.Time { return u.createdAt }

// UpdatedAt returns the last-updated timestamp.
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// --- Behavior ---

// Verify marks the user as verified.
func (u *User) Verify() {
	u.isVerified = true
	u.updatedAt = time.Now().UTC()
}

// UpdateProfile updates the user's profile information.
func (u *User) UpdateProfile(fullName, phone, avatarURL string) {
	if fullName != "" {
		u.fullName = fullName
	}
	if phone != "" {
		u.phone = Phone{value: phone}
	}
	if avatarURL != "" {
		u.avatarURL = avatarURL
	}
	u.updatedAt = time.Now().UTC()
}

// ChangePassword replaces the user's password hash.
func (u *User) ChangePassword(newHash string) {
	u.passwordHash = newHash
	u.updatedAt = time.Now().UTC()
}

// Deactivate deactivates the user by revoking verification.
func (u *User) Deactivate() {
	u.isVerified = false
	u.updatedAt = time.Now().UTC()
}

// IncrementVersion bumps the version for optimistic locking.
func (u *User) IncrementVersion() {
	u.version++
	u.updatedAt = time.Now().UTC()
}
