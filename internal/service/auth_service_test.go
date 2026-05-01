package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type userRepositoryMock struct {
	mock.Mock
}

func (m *userRepositoryMock) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *userRepositoryMock) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *userRepositoryMock) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *userRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *userRepositoryMock) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *userRepositoryMock) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *userRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type tokenServiceMock struct {
	mock.Mock
}

func (m *tokenServiceMock) GenerateTokens(userID uuid.UUID, email string, role string) (string, string, error) {
	args := m.Called(userID, email, role)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *tokenServiceMock) ValidateAccessToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	claims, _ := args.Get(0).(*Claims)
	return claims, args.Error(1)
}

func (m *tokenServiceMock) ValidateRefreshToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	claims, _ := args.Get(0).(*Claims)
	return claims, args.Error(1)
}

func (m *tokenServiceMock) RefreshTokens(refreshToken string) (string, string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func TestAuthServiceRegisterSuccess(t *testing.T) {
	ctx := context.Background()

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	req := &model.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	userRepo.On("GetByEmail", ctx, req.Email).Return((*model.User)(nil), repository.ErrUserNotFound).Once()
	userRepo.On("Create", ctx, mock.MatchedBy(func(user *model.User) bool {
		return user.Email == req.Email &&
			user.Name == req.Name &&
			user.Role == model.RoleUser &&
			user.Status == model.StatusActive &&
			user.PasswordHash != "" &&
			bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) == nil
	})).Return(nil).Once()

	user, err := svc.Register(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Name, user.Name)
	assert.Empty(t, user.PasswordHash)

	userRepo.AssertExpectations(t)
}

func TestAuthServiceRegisterDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	existing := &model.User{ID: uuid.New(), Email: "test@example.com"}

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(existing, nil).Once()

	user, err := svc.Register(ctx, &model.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, repository.ErrUserAlreadyExists)
	userRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	userRepo.AssertExpectations(t)
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(&model.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
		Role:         model.RoleUser,
	}, nil).Once()
	tokenSvc.On("GenerateTokens", userID, "test@example.com", string(model.RoleUser)).Return("access", "refresh", nil).Once()

	resp, err := svc.Login(ctx, &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "access", resp.AccessToken)
	assert.Equal(t, "refresh", resp.RefreshToken)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Empty(t, resp.User.PasswordHash)

	userRepo.AssertExpectations(t)
	tokenSvc.AssertExpectations(t)
}

func TestAuthServiceLoginInvalidPassword(t *testing.T) {
	ctx := context.Background()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(&model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         model.RoleUser,
	}, nil).Once()

	resp, err := svc.Login(ctx, &model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong-password",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	tokenSvc.AssertNotCalled(t, "GenerateTokens", mock.Anything, mock.Anything, mock.Anything)

	userRepo.AssertExpectations(t)
}

func TestAuthServiceUpdateProfileEmailConflict(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	userRepo.On("GetByID", ctx, userID).Return(&model.User{
		ID:    userID,
		Email: "old@example.com",
		Name:  "Old Name",
	}, nil).Once()
	userRepo.On("GetByEmail", ctx, "new@example.com").Return(&model.User{
		ID:    otherUserID,
		Email: "new@example.com",
	}, nil).Once()

	user, err := svc.UpdateProfile(ctx, userID, &model.UpdateUserRequest{
		Email: "new@example.com",
		Name:  "New Name",
	})

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, repository.ErrUserAlreadyExists)
	userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	userRepo.AssertExpectations(t)
}

func TestAuthServiceChangePasswordSuccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	currentPassword := "password123"
	newPassword := "newpassword123"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	userRepo.On("GetByID", ctx, userID).Return(&model.User{
		ID:           userID,
		PasswordHash: string(hashedPassword),
	}, nil).Once()
	userRepo.On("UpdatePassword", ctx, userID, mock.MatchedBy(func(hash string) bool {
		return hash != "" && bcrypt.CompareHashAndPassword([]byte(hash), []byte(newPassword)) == nil
	})).Return(nil).Once()

	err = svc.ChangePassword(ctx, userID, &model.ChangePasswordRequest{
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	})

	require.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthServiceChangePasswordSamePassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	password := "password123"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	userRepo.On("GetByID", ctx, userID).Return(&model.User{
		ID:           userID,
		PasswordHash: string(hashedPassword),
	}, nil).Once()

	err = svc.ChangePassword(ctx, userID, &model.ChangePasswordRequest{
		CurrentPassword: password,
		NewPassword:     password,
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSamePassword)
	userRepo.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	userRepo.AssertExpectations(t)
}

func TestAuthServiceRefreshTokenSuccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	tokenSvc.On("RefreshTokens", "refresh-token").Return("new-access", "new-refresh", nil).Once()
	tokenSvc.On("ValidateRefreshToken", "refresh-token").Return(&Claims{
		UserID: userID,
		Email:  "test@example.com",
		Role:   string(model.RoleUser),
	}, nil).Once()
	userRepo.On("GetByID", ctx, userID).Return(&model.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		Role:         model.RoleUser,
		PasswordHash: "secret",
	}, nil).Once()

	resp, err := svc.RefreshToken(ctx, "refresh-token")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "new-access", resp.AccessToken)
	assert.Equal(t, "new-refresh", resp.RefreshToken)
	assert.Equal(t, userID, resp.User.ID)
	assert.Empty(t, resp.User.PasswordHash)

	userRepo.AssertExpectations(t)
	tokenSvc.AssertExpectations(t)
}

func TestAuthServiceValidateEmail(t *testing.T) {
	ctx := context.Background()

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	userRepo.On("ExistsByEmail", ctx, "free@example.com").Return(false, nil).Once()
	userRepo.On("ExistsByEmail", ctx, "busy@example.com").Return(true, nil).Once()

	available, err := svc.ValidateEmail(ctx, "free@example.com")
	require.NoError(t, err)
	assert.True(t, available)

	available, err = svc.ValidateEmail(ctx, "busy@example.com")
	require.NoError(t, err)
	assert.False(t, available)

	userRepo.AssertExpectations(t)
}

func TestAuthServiceLoginTokenGenerationFailure(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(&model.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         model.RoleUser,
	}, nil).Once()
	tokenSvc.On("GenerateTokens", userID, "test@example.com", string(model.RoleUser)).Return("", "", errors.New("token error")).Once()

	resp, err := svc.Login(ctx, &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "failed to generate access token")

	userRepo.AssertExpectations(t)
	tokenSvc.AssertExpectations(t)
}

func TestAuthServiceGetProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	t.Run("Success", func(t *testing.T) {
		userRepo.On("GetByID", ctx, userID).Return(&model.User{
			ID:    userID,
			Email: "test@test.com",
		}, nil).Once()

		user, err := svc.GetProfile(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, userID, user.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		userRepo.On("GetByID", ctx, userID).Return((*model.User)(nil), repository.ErrUserNotFound).Once()

		user, err := svc.GetProfile(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestAuthServiceDeactivateProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	userRepo := new(userRepositoryMock)
	tokenSvc := new(tokenServiceMock)
	svc := NewAuthService(userRepo, tokenSvc)

	t.Run("Success", func(t *testing.T) {
		userRepo.On("GetByID", ctx, userID).Return(&model.User{ID: userID, Status: "active"}, nil).Once()
		userRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.Status == model.StatusBlocked
		})).Return(nil).Once()

		err := svc.DeactivateProfile(ctx, userID)
		assert.NoError(t, err)
	})
}
