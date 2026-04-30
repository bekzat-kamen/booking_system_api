package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminMiddlewareForbiddenForUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwt := new(tokenValidatorMock)
	router := gin.New()
	router.Use(AdminMiddleware(jwt))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	jwt.On("ValidateAccessToken", "user-token").Return(&service.Claims{
		UserID: uuid.New(),
		Email:  "user@example.com",
		Role:   string(model.RoleUser),
	}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin access required")
	jwt.AssertExpectations(t)
}

func TestAdminMiddlewareInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwt := new(tokenValidatorMock)
	router := gin.New()
	router.Use(AdminMiddleware(jwt))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	jwt.On("ValidateAccessToken", "bad-token").Return((*service.Claims)(nil), errors.New("bad token")).Once()

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
	jwt.AssertExpectations(t)
}

func TestOrganizerMiddlewareAllowsOrganizer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwt := new(tokenValidatorMock)
	router := gin.New()
	router.Use(OrganizerMiddleware(jwt))
	router.GET("/organizer", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"role": c.MustGet("user_role")})
	})

	jwt.On("ValidateAccessToken", "organizer-token").Return(&service.Claims{
		UserID: uuid.New(),
		Email:  "org@example.com",
		Role:   string(model.RoleOrganizer),
	}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/organizer", nil)
	req.Header.Set("Authorization", "Bearer organizer-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), string(model.RoleOrganizer))
	jwt.AssertExpectations(t)
}

func TestOrganizerMiddlewareRejectsGuest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwt := new(tokenValidatorMock)
	router := gin.New()
	router.Use(OrganizerMiddleware(jwt))
	router.GET("/organizer", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	jwt.On("ValidateAccessToken", "guest-token").Return(&service.Claims{
		UserID: uuid.New(),
		Email:  "guest@example.com",
		Role:   string(model.RoleGuest),
	}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/organizer", nil)
	req.Header.Set("Authorization", "Bearer guest-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "organizer or admin access required")
	jwt.AssertExpectations(t)
}
