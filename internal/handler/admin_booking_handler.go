package handler

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"

	_ "github.com/bekzat-kamen/booking_system_api/internal/model"
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

// GetAllBookings godoc
// @Summary [RU] Список всех бронирований / [EN] All bookings list
// @Description [RU] Возвращает список всех бронирований в системе с фильтрацией. / [EN] Returns a list of all bookings in the system with filtering.
// @Tags admin-bookings
// @Produce  json
// @Param page query int false "[RU] Страница / [EN] Page"
// @Param limit query int false "[RU] Лимит / [EN] Limit"
// @Param status query string false "[RU] Статус / [EN] Status"
// @Param user_id query string false "[RU] ID пользователя / [EN] User ID"
// @Param event_id query string false "[RU] ID мероприятия / [EN] Event ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/bookings [get]
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

// GetBookingDetail godoc
// @Summary [RU] Детали бронирования / [EN] Booking details
// @Description [RU] Возвращает полную информацию о бронировании, включая данные пользователя и билеты. / [EN] Returns full booking info, including user data and tickets.
// @Tags admin-bookings
// @Produce  json
// @Param id path string true "[RU] ID бронирования / [EN] Booking ID"
// @Success 200 {object} model.BookingDetail
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/bookings/{id} [get]
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

// CancelBooking godoc
// @Summary [RU] Отменить бронирование (админ) / [EN] Cancel booking (admin)
// @Description [RU] Принудительная отмена бронирования администратором. / [EN] Forced booking cancellation by admin.
// @Tags admin-bookings
// @Produce  json
// @Param id path string true "[RU] ID бронирования / [EN] Booking ID"
// @Param refund query bool false "[RU] Вернуть средства? / [EN] Refund money?"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /admin/bookings/{id}/cancel [post]
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

// RefundBooking godoc
// @Summary [RU] Оформить возврат / [EN] Process refund
// @Description [RU] Инициирует процесс возврата средств за бронирование. / [EN] Initiates the booking refund process.
// @Tags admin-bookings
// @Produce  json
// @Param id path string true "[RU] ID бронирования / [EN] Booking ID"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /admin/bookings/{id}/refund [post]
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

// GetBookingsStats godoc
// @Summary [RU] Статистика бронирований / [EN] Bookings statistics
// @Description [RU] Общая статистика по всем бронированиям в системе. / [EN] General statistics for all bookings in the system.
// @Tags admin-bookings
// @Produce  json
// @Success 200 {object} model.BookingsStats
// @Security BearerAuth
// @Router /admin/bookings/stats [get]
func (h *AdminBookingHandler) GetBookingsStats(c *gin.Context) {
	stats, err := h.adminBookingService.GetBookingsStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ExportBookings godoc
// @Summary [RU] Экспорт бронирований / [EN] Export bookings
// @Description [RU] Выгружает список бронирований в формате CSV. / [EN] Exports the booking list in CSV format.
// @Tags admin-bookings
// @Produce  text/csv
// @Param status query string false "[RU] Фильтр по статусу / [EN] Status filter"
// @Success 200 {file} file
// @Security BearerAuth
// @Router /admin/bookings/export [get]
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
