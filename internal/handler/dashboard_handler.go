package handler

import (
	"net/http"

	_ "github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService service.DashboardServiceInterface
}

func NewDashboardHandler(dashboardService service.DashboardServiceInterface) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetStats godoc
// @Summary [RU] Статистика дашборда / [EN] Dashboard statistics
// @Description [RU] Возвращает агрегированную статистику для панели администратора. / [EN] Returns aggregated statistics for the admin dashboard.
// @Tags admin
// @Produce  json
// @Success 200 {object} model.DashboardStats
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security BearerAuth
// @Router /admin/dashboard [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.dashboardService.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
