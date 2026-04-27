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
