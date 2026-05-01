package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

type authServiceMock struct {
	mock.Mock
}

func (m *authServiceMock) Register(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	args := m.Called(ctx, req)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *authServiceMock) Login(ctx context.Context, req *model.LoginRequest) (*service.LoginResponse, error) {
	args := m.Called(ctx, req)
	resp, _ := args.Get(0).(*service.LoginResponse)
	return resp, args.Error(1)
}

func (m *authServiceMock) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, userID)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *authServiceMock) UpdateProfile(ctx context.Context, userID uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	args := m.Called(ctx, userID, req)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *authServiceMock) ChangePassword(ctx context.Context, userID uuid.UUID, req *model.ChangePasswordRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *authServiceMock) DeactivateProfile(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *authServiceMock) ValidateEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *authServiceMock) RefreshToken(ctx context.Context, refreshToken string) (*service.LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	resp, _ := args.Get(0).(*service.LoginResponse)
	return resp, args.Error(1)
}

func performJSONRequest(r http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func addUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", uuid.MustParse("11111111-1111-1111-1111-111111111111"))
		c.Next()
	}
}

func TestAuthHandlerRegisterSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.POST("/register", h.Register)

	reqBody := model.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	authSvc.On("Register", mock.Anything, mock.MatchedBy(func(req *model.CreateUserRequest) bool {
		return req.Email == reqBody.Email && req.Name == reqBody.Name
	})).Return(&model.User{
		ID:    uuid.New(),
		Email: reqBody.Email,
		Name:  reqBody.Name,
		Role:  model.RoleUser,
	}, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/register", reqBody)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), reqBody.Email)
	authSvc.AssertExpectations(t)
}

func TestAuthHandlerRegisterConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.POST("/register", h.Register)

	authSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.CreateUserRequest")).
		Return((*model.User)(nil), repository.ErrUserAlreadyExists).Once()

	w := performJSONRequest(router, http.MethodPost, "/register", map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
	})

	require.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "email already exists")
	authSvc.AssertExpectations(t)
}

func TestAuthHandlerLoginUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.POST("/login", h.Login)

	authSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).
		Return((*service.LoginResponse)(nil), service.ErrInvalidCredentials).Once()

	w := performJSONRequest(router, http.MethodPost, "/login", map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	})

	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid email or password")
	authSvc.AssertExpectations(t)
}

func TestAuthHandlerGetProfileSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.Use(addUserContext())
	router.GET("/profile", h.GetProfile)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	authSvc.On("GetProfile", mock.Anything, userID).Return(&model.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
	}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/profile", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test@example.com")
	authSvc.AssertExpectations(t)
}

func TestAuthHandlerChangePasswordBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.Use(addUserContext())
	router.POST("/change-password", h.ChangePassword)

	w := performJSONRequest(router, http.MethodPost, "/change-password", map[string]string{
		"current_password": "",
		"new_password":     "short",
	})

	require.Equal(t, http.StatusBadRequest, w.Code)
	authSvc.AssertExpectations(t)
}

func TestAuthHandlerValidateEmailSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.GET("/validate-email", h.ValidateEmail)

	authSvc.On("ValidateEmail", mock.Anything, "free@example.com").Return(true, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/validate-email?email=free@example.com", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "true")
	authSvc.AssertExpectations(t)
}

func TestAuthHandlerUpdateProfileSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.Use(addUserContext())
	router.PUT("/profile", h.UpdateProfile)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	reqBody := model.UpdateUserRequest{Name: "New Name"}

	authSvc.On("UpdateProfile", mock.Anything, userID, &reqBody).Return(&model.User{ID: userID, Name: "New Name"}, nil).Once()

	w := performJSONRequest(router, http.MethodPut, "/profile", reqBody)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "New Name")
}

func TestAuthHandlerDeactivateProfileSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.Use(addUserContext())
	router.DELETE("/profile", h.DeactivateProfile)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	authSvc.On("DeactivateProfile", mock.Anything, userID).Return(nil).Once()

	w := performJSONRequest(router, http.MethodDelete, "/profile", nil)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandlerRefreshTokenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.POST("/refresh", h.RefreshToken)

	reqBody := map[string]string{"refresh_token": "old-refresh"}
	authSvc.On("RefreshToken", mock.Anything, "old-refresh").Return(&service.LoginResponse{AccessToken: "new-access"}, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/refresh", reqBody)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "new-access")
}

func TestAuthHandlerChangePasswordSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authSvc := new(authServiceMock)
	h := NewAuthHandler(authSvc)

	router := gin.New()
	router.Use(addUserContext())
	router.POST("/change-password", h.ChangePassword)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	reqBody := model.ChangePasswordRequest{CurrentPassword: "old", NewPassword: "newnewnew"}

	authSvc.On("ChangePassword", mock.Anything, userID, &reqBody).Return(nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/change-password", reqBody)

	require.Equal(t, http.StatusOK, w.Code)
}
