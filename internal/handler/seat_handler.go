package handler

import (
	"net/http"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SeatHandler struct {
	seatService service.SeatServiceInterface
}

func NewSeatHandler(seatService service.SeatServiceInterface) *SeatHandler {
	return &SeatHandler{seatService: seatService}
}

func (h *SeatHandler) GenerateSeats(c *gin.Context) {
	organizerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req model.GenerateSeatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.seatService.GenerateSeats(c.Request.Context(), eventID, organizerID.(uuid.UUID), &req)
	if err != nil {
		if err == service.ErrSeatsAlreadyGenerated {
			c.JSON(http.StatusConflict, gin.H{"error": "seats already generated for this event"})
			return
		}
		if err == repository.ErrEventNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		if err == service.ErrEventForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to manage seats for this event"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate seats"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "seats generated successfully"})
}

func (h *SeatHandler) GetSeatMap(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	seatMap, err := h.seatService.GetSeatMap(c.Request.Context(), eventID)
	if err != nil {
		if err == repository.ErrEventNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get seat map"})
		return
	}

	c.JSON(http.StatusOK, seatMap)
}

func (h *SeatHandler) GetAvailableSeats(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	seats, err := h.seatService.GetAvailableSeats(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get available seats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"seats": seats})
}
