package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type tokenValidatorMock struct {
	mock.Mock
}

func (m *tokenValidatorMock) ValidateAccessToken(tokenString string) (*service.Claims, error) {
	args := m.Called(tokenString)
	claims, _ := args.Get(0).(*service.Claims)
	return claims, args.Error(1)
}

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwt := new(tokenValidatorMock)
	router := gin.New()
	router.Use(AuthMiddleware(jwt))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header required")
	jwt.AssertExpectations(t)
}

func TestAuthMiddlewareSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwt := new(tokenValidatorMock)
	router := gin.New()
	router.Use(AuthMiddleware(jwt))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.MustGet("user_id"),
			"role":    c.MustGet("user_role"),
		})
	})

	userID := uuid.New()
	jwt.On("ValidateAccessToken", "valid-token").Return(&service.Claims{
		UserID: userID,
		Email:  "user@example.com",
		Role:   string(model.RoleUser),
	}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), userID.String())
	assert.Contains(t, w.Body.String(), string(model.RoleUser))
	jwt.AssertExpectations(t)
}
