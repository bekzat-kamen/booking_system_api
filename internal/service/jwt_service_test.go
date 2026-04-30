package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTServiceInvalidAccessDuration(t *testing.T) {
	svc, err := NewJWTService("access", "refresh", "bad", "24h")

	require.Error(t, err)
	assert.Nil(t, svc)
	assert.EqualError(t, err, "invalid access expire duration")
}

func TestJWTServiceGenerateAndValidateTokens(t *testing.T) {
	userID := uuid.New()
	svc, err := NewJWTService("access-secret", "refresh-secret", "15m", "24h")
	require.NoError(t, err)

	accessToken, refreshToken, err := svc.GenerateTokens(userID, "user@example.com", "user")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)

	accessClaims, err := svc.ValidateAccessToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, "user@example.com", accessClaims.Email)
	assert.Equal(t, "user", accessClaims.Role)

	refreshClaims, err := svc.ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
}

func TestJWTServiceValidateAccessTokenInvalid(t *testing.T) {
	svc, err := NewJWTService("access-secret", "refresh-secret", "15m", "24h")
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken("not-a-token")

	require.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTServiceValidateExpiredToken(t *testing.T) {
	userID := uuid.New()
	svc, err := NewJWTService("access-secret", "refresh-secret", "1ms", "24h")
	require.NoError(t, err)

	token, err := svc.generateToken(userID, "user@example.com", "user", svc.accessSecret, time.Millisecond)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	claims, err := svc.ValidateAccessToken(token)

	require.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestJWTServiceRefreshTokens(t *testing.T) {
	userID := uuid.New()
	svc, err := NewJWTService("access-secret", "refresh-secret", "15m", "24h")
	require.NoError(t, err)

	_, refreshToken, err := svc.GenerateTokens(userID, "user@example.com", "admin")
	require.NoError(t, err)

	newAccess, newRefresh, err := svc.RefreshTokens(refreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
}
