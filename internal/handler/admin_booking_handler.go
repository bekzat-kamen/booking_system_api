package handler

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"

	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminBookingHandler struct {
	adminBookingService service.AdminBookingServiceInterface
}

func NewAdminBookingHandler(adminBookingService service.AdminBookingServiceInterface) *AdminBookingHandler {
	return &AdminBookingHandler{adminBookingService: adminBookingService}
}

func (h *AdminBookingHandler) GetAllBookings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	userID := c.Query("user_id")
	eventID := c.Query("event_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	bookings, total, err := h.adminBookingService.GetAllBookings(c.Request.Context(), page, limit, status, userID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

func (h *AdminBookingHandler) GetBookingDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	detail, err := h.adminBookingService.GetBookingDetail(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get booking"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h *AdminBookingHandler) CancelBooking(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	refund, _ := strconv.ParseBool(c.DefaultQuery("refund", "false"))

	err = h.adminBookingService.CancelBooking(c.Request.Context(), id, refund)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if err.Error() == "cannot cancel this booking" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}

func (h *AdminBookingHandler) RefundBooking(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	err = h.adminBookingService.RefundBooking(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if errors.Is(err, repository.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		if err.Error() == "payment already refunded" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refund processed successfully"})
}

func (h *AdminBookingHandler) GetBookingsStats(c *gin.Context) {
	stats, err := h.adminBookingService.GetBookingsStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminBookingHandler) ExportBookings(c *gin.Context) {
	status := c.Query("status")

	rows, err := h.adminBookingService.ExportBookingsToCSV(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export bookings"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=bookings_export.csv")

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write csv"})
			return
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finalize csv"})
	}
}
