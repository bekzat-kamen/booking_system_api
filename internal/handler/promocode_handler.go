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

type PromocodeHandler struct {
	promocodeService service.PromocodeServiceInterface
}

func NewPromocodeHandler(promocodeService service.PromocodeServiceInterface) *PromocodeHandler {
	return &PromocodeHandler{promocodeService: promocodeService}
}

// CreatePromocode godoc
// @Summary [RU] Создать промокод / [EN] Create promocode
// @Description [RU] Создает новый промокод для скидок. / [EN] Creates a new promocode for discounts.
// @Tags promocodes
// @Accept  json
// @Produce  json
// @Param promocode body model.CreatePromocodeRequest true "[RU] Данные промокода / [EN] Promocode data"
// @Success 201 {object} model.PromocodeResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /promocodes [post]
func (h *PromocodeHandler) CreatePromocode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.CreatePromocodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promocode, err := h.promocodeService.CreatePromocode(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		if err == service.ErrInvalidDiscountValue {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create promocode"})
		return
	}

	c.JSON(http.StatusCreated, model.PromocodeResponse{
		ID:            promocode.ID,
		Code:          promocode.Code,
		Description:   promocode.Description,
		DiscountType:  promocode.DiscountType,
		DiscountValue: promocode.DiscountValue,
		MinAmount:     promocode.MinAmount,
		MaxUses:       promocode.MaxUses,
		UsedCount:     promocode.UsedCount,
		ValidFrom:     promocode.ValidFrom,
		ValidUntil:    promocode.ValidUntil,
		IsActive:      promocode.IsActive,
	})
}

// ValidatePromocode godoc
// @Summary [RU] Проверить промокод / [EN] Validate promocode
// @Description [RU] Проверяет валидность промокода для конкретного заказа. / [EN] Checks the validity of a promocode for a specific order.
// @Tags promocodes
// @Accept  json
// @Produce  json
// @Param request body model.ValidatePromocodeRequest true "[RU] Данные для проверки / [EN] Validation data"
// @Success 200 {object} model.PromocodeValidationResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /promocodes/validate [post]
func (h *PromocodeHandler) ValidatePromocode(c *gin.Context) {
	var req model.ValidatePromocodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.promocodeService.ValidatePromocode(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "validation failed"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPromocode godoc
// @Summary [RU] Получить промокод / [EN] Get promocode
// @Description [RU] Возвращает данные конкретного промокода по ID. / [EN] Returns specific promocode data by ID.
// @Tags promocodes
// @Produce  json
// @Param id path string true "[RU] ID промокода / [EN] Promocode ID"
// @Success 200 {object} model.PromocodeResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /promocodes/{id} [get]
func (h *PromocodeHandler) GetPromocode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promocode id"})
		return
	}

	promocode, err := h.promocodeService.GetPromocode(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPromocodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "promocode not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get promocode"})
		return
	}

	c.JSON(http.StatusOK, model.PromocodeResponse{
		ID:            promocode.ID,
		Code:          promocode.Code,
		Description:   promocode.Description,
		DiscountType:  promocode.DiscountType,
		DiscountValue: promocode.DiscountValue,
		MinAmount:     promocode.MinAmount,
		MaxUses:       promocode.MaxUses,
		UsedCount:     promocode.UsedCount,
		ValidFrom:     promocode.ValidFrom,
		ValidUntil:    promocode.ValidUntil,
		IsActive:      promocode.IsActive,
	})
}

// GetAllPromocodes godoc
// @Summary [RU] Список промокодов / [EN] List promocodes
// @Description [RU] Возвращает список всех доступных промокодов. / [EN] Returns a list of all available promocodes.
// @Tags promocodes
// @Produce  json
// @Param limit query int false "[RU] Лимит / [EN] Limit"
// @Param page query int false "[RU] Страница / [EN] Page"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /promocodes [get]
func (h *PromocodeHandler) GetAllPromocodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	promocodes, total, err := h.promocodeService.GetAllPromocodes(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get promocodes"})
		return
	}

	responses := make([]model.PromocodeResponse, 0, len(promocodes))
	for _, p := range promocodes {
		responses = append(responses, model.PromocodeResponse{
			ID:            p.ID,
			Code:          p.Code,
			Description:   p.Description,
			DiscountType:  p.DiscountType,
			DiscountValue: p.DiscountValue,
			MinAmount:     p.MinAmount,
			MaxUses:       p.MaxUses,
			UsedCount:     p.UsedCount,
			ValidFrom:     p.ValidFrom,
			ValidUntil:    p.ValidUntil,
			IsActive:      p.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"promocodes": responses,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// UpdatePromocode godoc
// @Summary [RU] Обновить промокод / [EN] Update promocode
// @Description [RU] Обновляет параметры существующего промокода. / [EN] Updates existing promocode parameters.
// @Tags promocodes
// @Accept  json
// @Produce  json
// @Param id path string true "[RU] ID промокода / [EN] Promocode ID"
// @Param promocode body model.UpdatePromocodeRequest true "[RU] Данные для обновления / [EN] Update data"
// @Success 200 {object} model.PromocodeResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /promocodes/{id} [put]
func (h *PromocodeHandler) UpdatePromocode(c *gin.Context) {
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

	promocode, err := h.promocodeService.UpdatePromocode(c.Request.Context(), id, &req)
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

	c.JSON(http.StatusOK, model.PromocodeResponse{
		ID:            promocode.ID,
		Code:          promocode.Code,
		Description:   promocode.Description,
		DiscountType:  promocode.DiscountType,
		DiscountValue: promocode.DiscountValue,
		MinAmount:     promocode.MinAmount,
		MaxUses:       promocode.MaxUses,
		UsedCount:     promocode.UsedCount,
		ValidFrom:     promocode.ValidFrom,
		ValidUntil:    promocode.ValidUntil,
		IsActive:      promocode.IsActive,
	})
}

// DeletePromocode godoc
// @Summary [RU] Удалить промокод / [EN] Delete promocode
// @Description [RU] Полностью удаляет промокод из системы. / [EN] Completely deletes a promocode from the system.
// @Tags promocodes
// @Produce  json
// @Param id path string true "[RU] ID промокода / [EN] Promocode ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /promocodes/{id} [delete]
func (h *PromocodeHandler) DeletePromocode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promocode id"})
		return
	}

	err = h.promocodeService.DeletePromocode(c.Request.Context(), id)
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

// DeactivatePromocode godoc
// @Summary [RU] Деактивировать промокод / [EN] Deactivate promocode
// @Description [RU] Делает промокод неактивным. / [EN] Makes the promocode inactive.
// @Tags promocodes
// @Produce  json
// @Param id path string true "[RU] ID промокода / [EN] Promocode ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /promocodes/{id}/deactivate [post]
func (h *PromocodeHandler) DeactivatePromocode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promocode id"})
		return
	}

	err = h.promocodeService.DeactivatePromocode(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPromocodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "promocode not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate promocode"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "promocode deactivated successfully"})
}
