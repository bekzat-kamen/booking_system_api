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

type BookingHandler struct {
	bookingService service.BookingServiceInterface
}

func NewBookingHandler(bookingService service.BookingServiceInterface) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// CreateBooking godoc
// @Summary [RU] Создать бронирование / [EN] Create booking
// @Description [RU] Бронирование мест на мероприятие. Резервирует места на время оплаты. / [EN] Booking seats for an event. Reserves seats for the duration of payment.
// @Tags bookings
// @Accept  json
// @Produce  json
// @Param booking body model.CreateBookingRequest true "[RU] Данные бронирования / [EN] Booking data"
// @Success 201 {object} model.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Security BearerAuth
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking, err := h.bookingService.CreateBooking(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		if err == repository.ErrEventNotFound || err == repository.ErrSeatNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrSeatsNotAvailable {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// GetBooking godoc
// @Summary [RU] Получить бронирование / [EN] Get booking
// @Description [RU] Возвращает детальную информацию о конкретном бронировании пользователя. / [EN] Returns detailed information about a specific user booking.
// @Tags bookings
// @Produce  json
// @Param id path string true "[RU] ID бронирования / [EN] Booking ID"
// @Success 200 {object} model.Booking
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	booking, err := h.bookingService.GetBooking(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get booking"})
		return
	}

	if booking.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized to view this booking"})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// GetUserBookings godoc
// @Summary [RU] Список бронирований пользователя / [EN] User bookings list
// @Description [RU] Возвращает историю бронирований текущего пользователя. / [EN] Returns the booking history of the current user.
// @Tags bookings
// @Produce  json
// @Param limit query int false "[RU] Лимит / [EN] Limit"
// @Param page query int false "[RU] Страница / [EN] Page"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /bookings [get]
func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	bookings, total, err := h.bookingService.GetUserBookings(c.Request.Context(), userID.(uuid.UUID), page, limit)
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

// CancelBooking godoc
// @Summary [RU] Отменить бронирование / [EN] Cancel booking
// @Description [RU] Отменяет активное бронирование до момента оплаты. / [EN] Cancels an active booking before payment.
// @Tags bookings
// @Produce  json
// @Param id path string true "[RU] ID бронирования / [EN] Booking ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	err = h.bookingService.CancelBooking(c.Request.Context(), id, userID.(uuid.UUID))
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if err == service.ErrCannotCancel {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}

// ConfirmBooking godoc
// @Summary [RU] Подтвердить бронирование / [EN] Confirm booking
// @Description [RU] Подтверждает бронирование (обычно вызывается системой после оплаты). / [EN] Confirms the booking (usually called by the system after payment).
// @Tags bookings
// @Produce  json
// @Param id path string true "[RU] ID бронирования / [EN] Booking ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 410 {object} map[string]string
// @Router /bookings/{id}/confirm [post]
func (h *BookingHandler) ConfirmBooking(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	err = h.bookingService.ConfirmBooking(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if err == service.ErrBookingExpired {
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking confirmed successfully"})
}
