package handler

import (
	"errors"
	"net/http"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService service.AuthServiceInterface
}
type LoginResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary [RU] Регистрация пользователя / [EN] User registration
// @Description [RU] Создает новый аккаунт пользователя. / [EN] Creates a new user account.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param user body model.CreateUserRequest true "[RU] Данные пользователя / [EN] User data"
// @Success 201 {object} model.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login godoc
// @Summary [RU] Вход в систему / [EN] User login
// @Description [RU] Аутентификация пользователя по email и паролю. / [EN] User authentication by email and password.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param credentials body model.LoginRequest true "[RU] Учетные данные / [EN] Credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// RefreshToken godoc
// @Summary [RU] Обновить токен / [EN] Refresh token
// @Description [RU] Получение нового access токена с помощью refresh токена. / [EN] Getting a new access token using a refresh token.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body object{refresh_token=string} true "[RU] Токен обновления / [EN] Refresh token"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrExpiredToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProfile godoc
// @Summary [RU] Профиль пользователя / [EN] User profile
// @Description [RU] Получение данных текущего авторизованного пользователя. / [EN] Getting current authorized user data.
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {object} model.User
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}
	user, err := h.authService.GetProfile(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary [RU] Обновить профиль / [EN] Update profile
// @Description [RU] Изменение персональных данных пользователя. / [EN] Changing user personal data.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param user body model.UpdateUserRequest true "[RU] Данные для обновления / [EN] Update data"
// @Success 200 {object} model.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Security BearerAuth
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePassword godoc
// @Summary [RU] Изменить пароль / [EN] Change password
// @Description [RU] Обновление пароля текущего пользователя. / [EN] Updating current user password.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body model.ChangePasswordRequest true "[RU] Запрос на смену пароля / [EN] Change password request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid current password"})
			return
		}
		if errors.Is(err, service.ErrSamePassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be different from current password"})
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// DeactivateProfile godoc
// @Summary [RU] Деактивировать профиль / [EN] Deactivate profile
// @Description [RU] Отключает аккаунт текущего пользователя. / [EN] Disables the current user account.
// @Tags auth
// @Produce  json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /auth/profile [delete]
func (h *AuthHandler) DeactivateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	if err := h.authService.DeactivateProfile(c.Request.Context(), userID.(uuid.UUID)); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deactivated successfully"})
}

// ValidateEmail godoc
// @Summary [RU] Проверить email / [EN] Validate email
// @Description [RU] Проверяет доступен ли email для регистрации. / [EN] Checks if an email is available for registration.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param email query string true "[RU] Email для проверки / [EN] Email to check"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/validate-email [get]
func (h *AuthHandler) ValidateEmail(c *gin.Context) {
	var req model.ValidateEmailRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	available, err := h.authService.ValidateEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available": available})
}
