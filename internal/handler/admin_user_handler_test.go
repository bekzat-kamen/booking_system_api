package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type adminUserServiceMock struct {
	mock.Mock
}

func (m *adminUserServiceMock) GetAllUsers(ctx context.Context, page, limit int, status, role string) ([]*model.User, int, error) {
	args := m.Called(ctx, page, limit, status, role)
	resp, _ := args.Get(0).([]*model.User)
	return resp, args.Int(1), args.Error(2)
}

func (m *adminUserServiceMock) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.User)
	return resp, args.Error(1)
}

func (m *adminUserServiceMock) GetUserDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(map[string]interface{})
	return resp, args.Error(1)
}

func (m *adminUserServiceMock) UpdateUserRole(ctx context.Context, id uuid.UUID, role model.Role, adminID uuid.UUID) error {
	args := m.Called(ctx, id, role, adminID)
	return args.Error(0)
}

func (m *adminUserServiceMock) BlockUser(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	args := m.Called(ctx, id, adminID)
	return args.Error(0)
}

func (m *adminUserServiceMock) UnblockUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *adminUserServiceMock) GetUserStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(map[string]int64)
	return resp, args.Error(1)
}

func (m *adminUserServiceMock) DeleteUser(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	args := m.Called(ctx, id, adminID)
	return args.Error(0)
}

func TestAdminUserHandlerGetAllUsersSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminUserServiceMock)
	h := NewAdminUserHandler(svc)
	router := gin.New()
	router.GET("/admin/users", h.GetAllUsers)

	svc.On("GetAllUsers", mock.Anything, 1, 20, "", "").Return([]*model.User{{ID: uuid.New(), Email: "user@example.com"}}, 1, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/users", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user@example.com")
}

func TestAdminUserHandlerGetUserDetailNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminUserServiceMock)
	h := NewAdminUserHandler(svc)
	router := gin.New()
	router.GET("/admin/users/:id", h.GetUserDetail)

	userID := uuid.New()
	svc.On("GetUserDetail", mock.Anything, userID).Return((map[string]interface{})(nil), repository.ErrUserNotFound).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/users/"+userID.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

func TestAdminUserHandlerUpdateUserRoleForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminUserServiceMock)
	h := NewAdminUserHandler(svc)
	router := gin.New()
	router.Use(addUserContext())
	router.PATCH("/admin/users/:id/role", h.UpdateUserRole)

	adminID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.New()
	svc.On("UpdateUserRole", mock.Anything, userID, model.RoleAdmin, adminID).Return(service.ErrCannotModifyAdmin).Once()

	w := performJSONRequest(router, http.MethodPatch, "/admin/users/"+userID.String()+"/role", map[string]string{
		"role": "admin",
	})

	require.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), service.ErrCannotModifyAdmin.Error())
}

func TestAdminUserHandlerGetUserStatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminUserServiceMock)
	h := NewAdminUserHandler(svc)
	router := gin.New()
	router.GET("/admin/users/stats", h.GetUserStats)

	svc.On("GetUserStats", mock.Anything).Return(map[string]int64{"total": 10}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/users/stats", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"total\":10")
}
