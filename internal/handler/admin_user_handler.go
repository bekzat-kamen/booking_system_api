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
	adminUserService *service.AdminUserService
}

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

func (h *AdminUserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.adminUserService.GetUserStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
