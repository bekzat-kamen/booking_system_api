package handler

import (
	"net/http"
	"strconv"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminUserHandler struct {
	adminUserService service.AdminUserServiceInterface
}

func NewAdminUserHandler(adminUserService service.AdminUserServiceInterface) *AdminUserHandler {
	return &AdminUserHandler{adminUserService: adminUserService}
}

// GetAllUsers godoc
// @Summary [RU] Список всех пользователей / [EN] All users list
// @Description [RU] Возвращает список всех пользователей системы с фильтрацией и пагинацией. / [EN] Returns a list of all system users with filtering and pagination.
// @Tags admin-users
// @Produce  json
// @Param page query int false "[RU] Страница / [EN] Page"
// @Param limit query int false "[RU] Лимит / [EN] Limit"
// @Param status query string false "[RU] Статус / [EN] Status"
// @Param role query string false "[RU] Роль / [EN] Role"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users [get]
func (h *AdminUserHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	role := c.Query("role")

	users, total, err := h.adminUserService.GetAllUsers(c.Request.Context(), page, limit, status, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}

	usersResponse := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		usersResponse = append(usersResponse, map[string]interface{}{
			"id":             user.ID,
			"email":          user.Email,
			"name":           user.Name,
			"role":           user.Role,
			"status":         user.Status,
			"email_verified": user.EmailVerified,
			"created_at":     user.CreatedAt,
			"updated_at":     user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": usersResponse,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetUserDetail godoc
// @Summary [RU] Детальная информация о пользователе / [EN] User details
// @Description [RU] Возвращает полную информацию о пользователе, включая его историю. / [EN] Returns full user information including history.
// @Tags admin-users
// @Produce  json
// @Param id path string true "[RU] ID пользователя / [EN] User ID"
// @Success 200 {object} model.UserDetail
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users/{id} [get]
func (h *AdminUserHandler) GetUserDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	detail, err := h.adminUserService.GetUserDetail(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// UpdateUserRole godoc
// @Summary [RU] Изменить роль пользователя / [EN] Update user role
// @Description [RU] Позволяет администратору изменить роль пользователя. / [EN] Allows admin to change user role.
// @Tags admin-users
// @Accept  json
// @Produce  json
// @Param id path string true "[RU] ID пользователя / [EN] User ID"
// @Param request body object{role=string} true "[RU] Новая роль / [EN] New role"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users/{id}/role [patch]
func (h *AdminUserHandler) UpdateUserRole(c *gin.Context) {
	adminID, _ := c.Get("user_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=user organizer moderator admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.adminUserService.UpdateUserRole(c.Request.Context(), id, model.Role(req.Role), adminID.(uuid.UUID))
	if err != nil {
		if err == service.ErrCannotModifyAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}

// BlockUser godoc
// @Summary [RU] Заблокировать пользователя / [EN] Block user
// @Description [RU] Переводит статус пользователя в BLOCKED. / [EN] Changes user status to BLOCKED.
// @Tags admin-users
// @Produce  json
// @Param id path string true "[RU] ID пользователя / [EN] User ID"
// @Success 200 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users/{id}/block [post]
func (h *AdminUserHandler) BlockUser(c *gin.Context) {
	adminID, _ := c.Get("user_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	err = h.adminUserService.BlockUser(c.Request.Context(), id, adminID.(uuid.UUID))
	if err != nil {
		if err == service.ErrCannotBlockYourself {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrCannotModifyAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user blocked successfully"})
}

// UnblockUser godoc
// @Summary [RU] Разблокировать пользователя / [EN] Unblock user
// @Description [RU] Восстанавливает активный статус пользователя. / [EN] Restores active user status.
// @Tags admin-users
// @Produce  json
// @Param id path string true "[RU] ID пользователя / [EN] User ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users/{id}/unblock [post]
func (h *AdminUserHandler) UnblockUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	err = h.adminUserService.UnblockUser(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unblock user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user unblocked successfully"})
}

// GetUserStats godoc
// @Summary [RU] Статистика пользователей / [EN] User stats
// @Description [RU] Возвращает статистику по количеству и статусам пользователей. / [EN] Returns statistics on user count and statuses.
// @Tags admin-users
// @Produce  json
// @Success 200 {object} model.UserStats
// @Security BearerAuth
// @Router /admin/users/stats [get]
func (h *AdminUserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.adminUserService.GetUserStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
