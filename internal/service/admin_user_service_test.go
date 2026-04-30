package service

import (
	"context"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type adminUserRepositoryMock struct {
	mock.Mock
}

func (m *adminUserRepositoryMock) GetAllUsers(ctx context.Context, page, limit int, status, role string) ([]*model.User, int, error) {
	args := m.Called(ctx, page, limit, status, role)
	resp, _ := args.Get(0).([]*model.User)
	return resp, args.Int(1), args.Error(2)
}

func (m *adminUserRepositoryMock) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.User)
	return resp, args.Error(1)
}

func (m *adminUserRepositoryMock) UpdateUserRole(ctx context.Context, id uuid.UUID, role model.Role) error {
	args := m.Called(ctx, id, role)
	return args.Error(0)
}

func (m *adminUserRepositoryMock) UpdateUserStatus(ctx context.Context, id uuid.UUID, status model.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *adminUserRepositoryMock) GetUserStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(map[string]int64)
	return resp, args.Error(1)
}

func (m *adminUserRepositoryMock) GetUserBookingsCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *adminUserRepositoryMock) GetUserSpentAmount(ctx context.Context, userID uuid.UUID) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func TestAdminUserServiceGetUserByIDClearsPassword(t *testing.T) {
	ctx := context.Background()
	repo := new(adminUserRepositoryMock)
	svc := NewAdminUserService(repo)
	userID := uuid.New()

	repo.On("GetUserByID", ctx, userID).Return(&model.User{ID: userID, PasswordHash: "secret"}, nil).Once()

	user, err := svc.GetUserByID(ctx, userID)

	require.NoError(t, err)
	assert.Empty(t, user.PasswordHash)
}

func TestAdminUserServiceUpdateUserRoleCannotModifyAdmin(t *testing.T) {
	ctx := context.Background()
	repo := new(adminUserRepositoryMock)
	svc := NewAdminUserService(repo)
	userID := uuid.New()
	adminID := uuid.New()

	repo.On("GetUserByID", ctx, userID).Return(&model.User{ID: userID, Role: model.RoleAdmin}, nil).Once()

	err := svc.UpdateUserRole(ctx, userID, model.RoleUser, adminID)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotModifyAdmin)
}

func TestAdminUserServiceBlockUserCannotBlockYourself(t *testing.T) {
	svc := NewAdminUserService(new(adminUserRepositoryMock))
	adminID := uuid.New()

	err := svc.BlockUser(context.Background(), adminID, adminID)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotBlockYourself)
}

func TestAdminUserServiceDeleteUserSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(adminUserRepositoryMock)
	svc := NewAdminUserService(repo)
	userID := uuid.New()
	adminID := uuid.New()

	repo.On("GetUserByID", ctx, userID).Return(&model.User{ID: userID, Role: model.RoleUser}, nil).Once()
	repo.On("UpdateUserStatus", ctx, userID, model.StatusBlocked).Return(nil).Once()

	err := svc.DeleteUser(ctx, userID, adminID)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}
