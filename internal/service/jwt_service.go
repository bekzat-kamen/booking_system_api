package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// JWTService работает с JWT токенами
type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
}

func NewJWTService(accessSecret, refreshSecret string, accessExpire, refreshExpire string) (*JWTService, error) {
	accessDur, err := time.ParseDuration(accessExpire)
	if err != nil {
		return nil, errors.New("invalid access expire duration")
	}

	refreshDur, err := time.ParseDuration(refreshExpire)
	if err != nil {
		return nil, errors.New("invalid refresh expire duration")
	}

	return &JWTService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpire:  accessDur,
		refreshExpire: refreshDur,
	}, nil
}

func (s *JWTService) GenerateTokens(userID uuid.UUID, email string, role string) (string, string, error) {
	// Генерируем Access Token
	accessToken, err := s.generateToken(userID, email, role, s.accessSecret, s.accessExpire)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateToken(userID, email, role, s.refreshSecret, s.refreshExpire)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *JWTService) generateToken(userID uuid.UUID, email string, role string, secret []byte, expire time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.accessSecret)
}

func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.refreshSecret)
}

func (s *JWTService) validateToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *JWTService) RefreshTokens(refreshToken string) (string, string, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	return s.GenerateTokens(claims.UserID, claims.Email, claims.Role)
}
