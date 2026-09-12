package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/auth"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/domain"
	"github.com/niaga-labs/niaga-labs-pet-lib-proto/dto"
	"github.com/niaga-labs/niaga-labs-pet-service-identity/internal/domain/identity"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=owner runner admin shop"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the response for authentication operations.
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         dto.UserDTO `json:"user"`
}

// UpdateProfileRequest represents a profile update request.
type UpdateProfileRequest struct {
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url"`
}

// AuthService implements authentication and user management use cases.
type AuthService struct {
	userRepo          identity.UserRepository
	tokenRepo         identity.TokenRepository
	passwordResetRepo identity.PasswordResetRepository
	notifier          PasswordResetNotifier
	jwt               *auth.JWTManager
	logger            *zap.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo identity.UserRepository,
	tokenRepo identity.TokenRepository,
	passwordResetRepo identity.PasswordResetRepository,
	notifier PasswordResetNotifier,
	jwt *auth.JWTManager,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		tokenRepo:         tokenRepo,
		passwordResetRepo: passwordResetRepo,
		notifier:          notifier,
		jwt:               jwt,
		logger:            logger,
	}
}

// Register creates a new user account and returns authentication tokens.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if email is already taken
	existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, domain.NewAlreadyExistsError("User", "email", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create domain user
	role := auth.UserRole(req.Role)
	user, err := identity.NewUser(req.Email, req.Phone, req.FullName, string(hashedPassword), role)
	if err != nil {
		return nil, domain.NewValidationError(err.Error())
	}

	// Persist user
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user", zap.Error(err))
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(user.ID(), user.Email(), user.Role())
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenStr, err := s.jwt.GenerateRefreshToken(user.ID())
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	refreshToken := identity.NewRefreshToken(user.ID(), refreshTokenStr, time.Now().Add(7*24*time.Hour))
	if err := s.tokenRepo.Save(ctx, refreshToken); err != nil {
		s.logger.Error("failed to save refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	s.logger.Info("user registered", zap.String("user_id", user.ID().String()), zap.String("email", user.Email()))

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         toUserDTO(user),
	}, nil
}

// Login authenticates a user by email and password.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.NewUnauthorizedError("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash()), []byte(req.Password)); err != nil {
		return nil, domain.NewUnauthorizedError("invalid email or password")
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(user.ID(), user.Email(), user.Role())
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenStr, err := s.jwt.GenerateRefreshToken(user.ID())
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	refreshToken := identity.NewRefreshToken(user.ID(), refreshTokenStr, time.Now().Add(7*24*time.Hour))
	if err := s.tokenRepo.Save(ctx, refreshToken); err != nil {
		s.logger.Error("failed to save refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	s.logger.Info("user logged in", zap.String("user_id", user.ID().String()), zap.String("email", user.Email()))

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         toUserDTO(user),
	}, nil
}

// RefreshToken validates a refresh token and issues a new token pair.
func (s *AuthService) RefreshToken(ctx context.Context, token string) (*AuthResponse, error) {
	// Validate the JWT signature of the refresh token
	claims, err := s.jwt.ValidateToken(token)
	if err != nil {
		return nil, domain.NewUnauthorizedError("invalid refresh token")
	}

	// Lookup stored refresh token
	storedToken, err := s.tokenRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, domain.NewUnauthorizedError("refresh token not found")
	}

	if !storedToken.IsValid() {
		return nil, domain.NewUnauthorizedError("refresh token is expired or revoked")
	}

	// Revoke the old token
	storedToken.Revoke()

	// Find the user
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.NewNotFoundError("User", claims.UserID.String())
	}

	// Generate new token pair
	accessToken, err := s.jwt.GenerateAccessToken(user.ID(), user.Email(), user.Role())
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenStr, err := s.jwt.GenerateRefreshToken(user.ID())
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store new refresh token
	newRefreshToken := identity.NewRefreshToken(user.ID(), refreshTokenStr, time.Now().Add(7*24*time.Hour))
	if err := s.tokenRepo.Save(ctx, newRefreshToken); err != nil {
		s.logger.Error("failed to save refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	s.logger.Info("token refreshed", zap.String("user_id", user.ID().String()))

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         toUserDTO(user),
	}, nil
}

// Logout revokes all refresh tokens for the specified user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.tokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		s.logger.Error("failed to revoke tokens", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	s.logger.Info("user logged out", zap.String("user_id", userID.String()))
	return nil
}

// GetProfile retrieves the user profile by ID.
func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.NewNotFoundError("User", userID.String())
	}

	result := toUserDTO(user)
	return &result, nil
}

// UpdateProfile updates the user's profile information.
func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.NewNotFoundError("User", userID.String())
	}

	user.UpdateProfile(req.FullName, req.Phone, req.AvatarURL)
	user.IncrementVersion()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user", zap.Error(err))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("user profile updated", zap.String("user_id", userID.String()))

	result := toUserDTO(user)
	return &result, nil
}

// --- Admin methods ---

// UserStatsDTO holds user statistics for the admin dashboard.
type UserStatsDTO struct {
	TotalUsers int64            `json:"total_users"`
	ByRole     map[string]int64 `json:"by_role"`
}

// ListUsers returns a paginated list of all users.
func (s *AuthService) ListUsers(ctx context.Context, page, limit int) ([]dto.UserDTO, int64, error) {
	users, total, err := s.userRepo.ListAll(ctx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	dtos := make([]dto.UserDTO, len(users))
	for i, u := range users {
		dtos[i] = toUserDTO(u)
	}
	return dtos, total, nil
}

// GetUserByID retrieves a single user by ID (admin).
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.NewNotFoundError("User", userID.String())
	}
	result := toUserDTO(user)
	return &result, nil
}

// BanUser deactivates a user account.
func (s *AuthService) BanUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return domain.NewNotFoundError("User", userID.String())
	}

	user.Deactivate()
	user.IncrementVersion()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to ban user", zap.Error(err))
		return fmt.Errorf("failed to ban user: %w", err)
	}

	// Revoke all tokens so user is logged out
	_ = s.tokenRepo.RevokeAllForUser(ctx, userID)

	s.logger.Info("user banned", zap.String("user_id", userID.String()))
	return nil
}

// GetUserStats returns aggregate user statistics.
func (s *AuthService) GetUserStats(ctx context.Context) (*UserStatsDTO, error) {
	counts, err := s.userRepo.CountByRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	var total int64
	for _, c := range counts {
		total += c
	}

	return &UserStatsDTO{
		TotalUsers: total,
		ByRole:     counts,
	}, nil
}

// ResetPassword validates a reset token and updates the user's password hash atomically.
func (s *AuthService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	if len(req.NewPassword) < 8 {
		return domain.NewValidationError("password must be at least 8 characters")
	}

	reset, err := s.passwordResetRepo.FindAnyByToken(ctx, req.Token)
	if err != nil {
		return domain.NewValidationError("invalid token")
	}

	if reset.IsUsed() {
		return domain.NewValidationError("token already used")
	}

	if reset.IsExpired() {
		return domain.NewValidationError("token expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password during reset", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.passwordResetRepo.MarkUsedAndUpdatePassword(ctx, reset.ID(), reset.UserID(), string(hashedPassword)); err != nil {
		s.logger.Error("failed to complete password reset transaction", zap.Error(err))
		return fmt.Errorf("failed to reset password: %w", err)
	}

	s.logger.Info("password reset completed", zap.String("user_id", reset.UserID().String()))
	return nil
}

// ForgotPassword initiates a password reset for the given email.
// Always returns nil so the handler can respond 202 regardless of whether the email exists.
func (s *AuthService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error {
	if req.Email == "" {
		return nil
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.logger.Error("failed to generate reset token", zap.Error(err))
		return fmt.Errorf("failed to generate reset token: %w", err)
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	reset := identity.NewPasswordReset(user.ID(), tokenStr, time.Now().UTC().Add(time.Hour))
	if err := s.passwordResetRepo.Create(ctx, reset); err != nil {
		s.logger.Error("failed to persist password reset token", zap.Error(err))
		return fmt.Errorf("failed to persist password reset token: %w", err)
	}

	if err := s.notifier.SendPasswordResetEmail(ctx, user.Email(), tokenStr); err != nil {
		s.logger.Warn("failed to enqueue password reset email", zap.Error(err), zap.String("email", user.Email()))
	}

	return nil
}

// toUserDTO converts a domain User to a UserDTO.
func toUserDTO(user *identity.User) dto.UserDTO {
	return dto.UserDTO{
		ID:         user.ID(),
		Email:      user.Email(),
		Phone:      user.Phone(),
		FullName:   user.FullName(),
		Role:       string(user.Role()),
		IsVerified: user.IsVerified(),
		AvatarURL:  user.AvatarURL(),
		CreatedAt:  user.CreatedAt(),
	}
}
