package service

import (
	"context"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidPassword    = errors.New("invalid current password")
	ErrSamePassword       = errors.New("new password must be different from current password")
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtService *JWTService
}

type LoginResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

func NewAuthService(userRepo *repository.UserRepository, jwtService *JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *AuthService) Register(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {

	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, repository.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		ID:            uuid.New(),
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		Name:          req.Name,
		Role:          model.RoleUser,
		Status:        model.StatusActive,
		EmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*LoginResponse, error) {

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokens(
		user.ID,
		user.Email,
		string(user.Role),
	)

	if err != nil {
		return nil, errors.New("failed to generate access token")
	}
	user.PasswordHash = ""
	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Email != "" && req.Email != user.Email {
		existing, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err == nil && existing != nil && existing.ID != user.ID {
			return nil, repository.ErrUserAlreadyExists
		}
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, err
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req *model.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidPassword
	}

	if req.CurrentPassword == req.NewPassword {
		return ErrSamePassword
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(newHashedPassword))
}

func (s *AuthService) DeactivateProfile(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Status = model.StatusBlocked
	return s.userRepo.Update(ctx, user)
}

func (s *AuthService) ValidateEmail(ctx context.Context, email string) (bool, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	newAccess, newRefresh, err := s.jwtService.RefreshTokens(refreshToken)
	if err != nil {
		return nil, err
	}

	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	user.PasswordHash = ""
	return &LoginResponse{
		User:         user,
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}, nil
}
