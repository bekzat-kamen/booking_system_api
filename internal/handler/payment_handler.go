package handler

import (
	"net/http"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService service.PaymentServiceInterface
}

func NewPaymentHandler(paymentService service.PaymentServiceInterface) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// CreatePayment godoc
// @Summary [RU] Создать платеж / [EN] Create payment
// @Description [RU] Создает новую запись о платеже для бронирования. / [EN] Creates a new payment record for a booking.
// @Tags payments
// @Accept  json
// @Produce  json
// @Param payment body model.CreatePaymentRequest true "[RU] Данные платежа / [EN] Payment data"
// @Success 201 {object} model.Payment
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Security BearerAuth
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req model.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if err == service.ErrPaymentAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// ProcessPayment godoc
// @Summary [RU] Обработать платеж / [EN] Process payment
// @Description [RU] Имитирует процесс оплаты (переводит в статус SUCCESS). / [EN] Simulates the payment process (changes status to SUCCESS).
// @Tags payments
// @Produce  json
// @Param id path string true "[RU] ID платежа / [EN] Payment ID"
// @Success 200 {object} model.Payment
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /payments/{id}/process [post]
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	payment, err := h.paymentService.ProcessPayment(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		if err == service.ErrInvalidPaymentStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetPayment godoc
// @Summary [RU] Получить информацию о платеже / [EN] Get payment info
// @Description [RU] Возвращает данные конкретного платежа. / [EN] Returns specific payment data.
// @Tags payments
// @Produce  json
// @Param id path string true "[RU] ID платежа / [EN] Payment ID"
// @Success 200 {object} model.Payment
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	payment, err := h.paymentService.GetPayment(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetPaymentByBooking godoc
// @Summary [RU] Платеж по ID бронирования / [EN] Get payment by booking ID
// @Description [RU] Возвращает данные платежа, связанного с указанным бронированием. / [EN] Returns payment data associated with the specified booking.
// @Tags payments
// @Produce  json
// @Param booking_id path string true "[RU] ID бронирования / [EN] Booking ID"
// @Success 200 {object} model.Payment
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /payments/booking/{booking_id} [get]
func (h *PaymentHandler) GetPaymentByBooking(c *gin.Context) {
	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	payment, err := h.paymentService.GetPaymentByBooking(c.Request.Context(), bookingID)
	if err != nil {
		if err == repository.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// Webhook godoc
// @Summary [RU] Вебхук платежной системы / [EN] Payment system webhook
// @Description [RU] Конечная точка для приема уведомлений от платежной системы. / [EN] Endpoint for receiving notifications from the payment system.
// @Tags payments
// @Accept  json
// @Produce  json
// @Param webhook body model.WebhookRequest true "[RU] Данные вебхука / [EN] Webhook data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /payments/webhook [post]
func (h *PaymentHandler) Webhook(c *gin.Context) {
	var req model.WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.paymentService.ProcessWebhook(c.Request.Context(), &req)
	if err != nil {
		if err == repository.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook processed successfully"})
}
