package handler

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminPromocodeHandler struct {
	adminPromocodeService service.AdminPromocodeServiceInterface
}

func NewAdminPromocodeHandler(adminPromocodeService service.AdminPromocodeServiceInterface) *AdminPromocodeHandler {
	return &AdminPromocodeHandler{adminPromocodeService: adminPromocodeService}
}

// GetAllPromocodes godoc
// @Summary [RU] Список всех промокодов / [EN] All promocodes list
// @Description [RU] Возвращает полный список промокодов для администратора. / [EN] Returns full list of promocodes for admin.
// @Tags admin-promocodes
// @Produce  json
// @Param page query int false "[RU] Страница / [EN] Page"
// @Param limit query int false "[RU] Лимит / [EN] Limit"
// @Param is_active query string false "[RU] Активен (true/false) / [EN] Is active (true/false)"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/promocodes [get]
func (h *AdminPromocodeHandler) GetAllPromocodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	isActive := c.Query("is_active")

	promocodes, total, err := h.adminPromocodeService.GetAllPromocodes(c.Request.Context(), page, limit, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get promocodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"promocodes": promocodes,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetPromocodeDetail godoc
// @Summary [RU] Детали промокода / [EN] Promocode details
// @Description [RU] Возвращает детальную информацию о промокоде, включая историю использования. / [EN] Returns detailed promocode info, including usage history.
// @Tags admin-promocodes
// @Produce  json
// @Param id path string true "[RU] ID промокода / [EN] Promocode ID"
// @Success 200 {object} model.PromocodeDetail
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/promocodes/{id} [get]
func (h *AdminPromocodeHandler) GetPromocodeDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promocode id"})
		return
	}

	detail, err := h.adminPromocodeService.GetPromocodeDetail(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPromocodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "promocode not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get promocode"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// UpdatePromocode godoc
// @Summary [RU] Обновить промокод (админ) / [EN] Update promocode (admin)
// @Description [RU] Принудительное обновление параметров промокода. / [EN] Forced update of promocode parameters.
// @Tags admin-promocodes
// @Accept  json
// @Produce  json
// @Param id path string true "[RU] ID промокода / [EN] Promocode ID"
// @Param promocode body model.UpdatePromocodeRequest true "[RU] Данные для обновления / [EN] Update data"
// @Success 200 {object} model.Promocode
// @Security BearerAuth
// @Router /admin/promocodes/{id} [put]
func (h *AdminPromocodeHandler) UpdatePromocode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promocode id"})
		return
	}

	var req model.UpdatePromocodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promocode, err := h.adminPromocodeService.UpdatePromocode(c.Request.Context(), id, &req)
	if err != nil {
		if err == repository.ErrPromocodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "promocode not found"})
			return
		}
		if err == service.ErrInvalidDiscountValue {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update promocode"})
		return
	}

	c.JSON(http.StatusOK, promocode)
}

// DeletePromocode godoc
// @Summary [RU] Удалить промокод (админ) / [EN] Delete promocode (admin)
// @Description [RU] Удаление промокода из базы данных администратором. / [EN] Deletion of a promocode from the database by admin.
// @Tags admin-promocodes
// @Produce  json
// @Param id path string true "[RU] ID промокода / [EN] Promocode ID"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /admin/promocodes/{id} [delete]
func (h *AdminPromocodeHandler) DeletePromocode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promocode id"})
		return
	}

	err = h.adminPromocodeService.DeletePromocode(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPromocodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "promocode not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete promocode"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "promocode deleted successfully"})
}

// BulkDeactivate godoc
// @Summary [RU] Массовая деактивация / [EN] Bulk deactivate
// @Description [RU] Деактивация сразу нескольких промокодов по их ID. / [EN] Deactivation of several promocodes by their IDs at once.
// @Tags admin-promocodes
// @Accept  json
// @Produce  json
// @Param request body object{ids=[]string} true "[RU] Список ID / [EN] List of IDs"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /admin/promocodes/bulk-deactivate [post]
func (h *AdminPromocodeHandler) BulkDeactivate(c *gin.Context) {
	var req struct {
		Ids []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ids := make([]uuid.UUID, 0, len(req.Ids))
	for _, idStr := range req.Ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid: " + idStr})
			return
		}
		ids = append(ids, id)
	}

	err := h.adminPromocodeService.BulkDeactivate(c.Request.Context(), ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate promocodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "promocodes deactivated successfully"})
}

// GetPromocodesStats godoc
// @Summary [RU] Статистика промокодов / [EN] Promocodes statistics
// @Description [RU] Общая статистика по использованию промокодов. / [EN] General statistics on promocode usage.
// @Tags admin-promocodes
// @Produce  json
// @Success 200 {object} model.PromocodesStats
// @Security BearerAuth
// @Router /admin/promocodes/stats [get]
func (h *AdminPromocodeHandler) GetPromocodesStats(c *gin.Context) {
	stats, err := h.adminPromocodeService.GetPromocodesStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ExportPromocodes godoc
// @Summary [RU] Экспорт промокодов / [EN] Export promocodes
// @Description [RU] Выгружает список всех промокодов в CSV. / [EN] Exports the list of all promocodes to CSV.
// @Tags admin-promocodes
// @Produce  text/csv
// @Success 200 {file} file
// @Security BearerAuth
// @Router /admin/promocodes/export [get]
func (h *AdminPromocodeHandler) ExportPromocodes(c *gin.Context) {
	rows, err := h.adminPromocodeService.ExportPromocodesToCSV(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export promocodes"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=promocodes_export.csv")

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write csv"})
			return
		}
	}

	if err := w.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finalize csv"})
		return
	}
}
