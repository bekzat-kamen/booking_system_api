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
	adminPromocodeService *service.AdminPromocodeService
}

func NewAdminPromocodeHandler(adminPromocodeService *service.AdminPromocodeService) *AdminPromocodeHandler {
	return &AdminPromocodeHandler{adminPromocodeService: adminPromocodeService}
}

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

func (h *AdminPromocodeHandler) GetPromocodesStats(c *gin.Context) {
	stats, err := h.adminPromocodeService.GetPromocodesStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

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
		w.Write(row)
	}
}
